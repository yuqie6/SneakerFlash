package logger

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"SneakerFlash/internal/config"

	"gopkg.in/natefinch/lumberjack.v2"
)

func InitLogger(cfg config.LoggerConfig, service string) {
	filePath := resolveLogPath(cfg.Path, service)
	if err := os.MkdirAll(filepath.Dir(filePath), 0o755); err != nil {
		panic(fmt.Sprintf("创建日志目录失败: %v", err))
	}

	fileWriter := &lumberjack.Logger{
		Filename:   filePath,
		MaxSize:    cfg.MaxSize,
		MaxAge:     cfg.MaxAge,
		MaxBackups: cfg.MaxBackups,
	}

	var loggerLevel slog.Level
	switch strings.ToLower(cfg.Level) {
	case "info":
		loggerLevel = slog.LevelInfo
	case "debug":
		loggerLevel = slog.LevelDebug
	case "warn":
		loggerLevel = slog.LevelWarn
	case "error":
		loggerLevel = slog.LevelError
	default:
		loggerLevel = slog.LevelInfo
	}

	jsonHandler := slog.NewJSONHandler(fileWriter, &slog.HandlerOptions{
		Level:     loggerLevel,
		AddSource: true,
	})

	if loggerLevel == slog.LevelDebug {
		consoleHandler := newConsoleHandler(os.Stdout, &slog.HandlerOptions{
			Level:     loggerLevel,
			AddSource: true,
		})
		slog.SetDefault(slog.New(newMultiHandler(jsonHandler, consoleHandler)))
		return
	}

	slog.SetDefault(slog.New(jsonHandler))
}

type multiHandler struct {
	handlers []slog.Handler
}

func newMultiHandler(handlers ...slog.Handler) slog.Handler {
	return &multiHandler{handlers: handlers}
}

func (m *multiHandler) Enabled(ctx context.Context, level slog.Level) bool {
	// 任一 handler 支持该级别即允许日志通过
	for _, h := range m.handlers {
		if h.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

func (m *multiHandler) Handle(ctx context.Context, r slog.Record) error {
	// 逐个 handler 写入，错误优先返回首个
	var firstErr error
	for _, h := range m.handlers {
		if !h.Enabled(ctx, r.Level) {
			continue
		}
		if err := h.Handle(ctx, r.Clone()); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

func (m *multiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	handlers := make([]slog.Handler, 0, len(m.handlers))
	for _, h := range m.handlers {
		handlers = append(handlers, h.WithAttrs(attrs))
	}
	return &multiHandler{handlers: handlers}
}

func (m *multiHandler) WithGroup(name string) slog.Handler {
	handlers := make([]slog.Handler, 0, len(m.handlers))
	for _, h := range m.handlers {
		handlers = append(handlers, h.WithGroup(name))
	}
	return &multiHandler{handlers: handlers}
}

type consoleHandler struct {
	writer      io.Writer
	level       slog.Level
	addSource   bool
	attrs       []slog.Attr
	groups      []string
	replaceAttr func([]string, slog.Attr) slog.Attr
}

func newConsoleHandler(w io.Writer, opts *slog.HandlerOptions) slog.Handler {
	if opts == nil {
		opts = &slog.HandlerOptions{}
	}
	level := slog.LevelInfo
	if opts.Level != nil {
		level = opts.Level.Level()
	}
	return &consoleHandler{
		writer:      w,
		level:       level,
		addSource:   opts.AddSource,
		replaceAttr: opts.ReplaceAttr,
	}
}

func (h *consoleHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level
}

func (h *consoleHandler) Handle(ctx context.Context, r slog.Record) error {
	if !h.Enabled(ctx, r.Level) {
		return nil
	}

	var b strings.Builder
	timestamp := r.Time
	if timestamp.IsZero() {
		timestamp = time.Now()
	}
	fmt.Fprintf(&b, "[%s] %s", strings.ToUpper(r.Level.String()), timestamp.Format("2006-01-02 15:04:05.000"))
	if r.Message != "" {
		b.WriteByte(' ')
		b.WriteString(r.Message)
	}

	for _, attr := range h.attrs {
		h.writeAttr(&b, h.groups, attr)
	}
	r.Attrs(func(attr slog.Attr) bool {
		h.writeAttr(&b, h.groups, attr)
		return true
	})

	if h.addSource {
		src := r.Source()
		if src.File != "" {
			fmt.Fprintf(&b, " source=%s:%d", src.File, src.Line)
		}
	}

	b.WriteByte('\n')
	_, err := io.WriteString(h.writer, b.String())
	return err
}

func (h *consoleHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	cp := *h
	cp.attrs = append(append([]slog.Attr{}, h.attrs...), attrs...)
	return &cp
}

func (h *consoleHandler) WithGroup(name string) slog.Handler {
	cp := *h
	cp.groups = append(append([]string{}, h.groups...), name)
	return &cp
}

func (h *consoleHandler) writeAttr(b *strings.Builder, groups []string, attr slog.Attr) {
	if h.replaceAttr != nil {
		attr = h.replaceAttr(groups, attr)
	}
	if attr.Equal(slog.Attr{}) {
		return
	}

	attr.Value = attr.Value.Resolve()
	if attr.Value.Kind() == slog.KindGroup {
		childGroups := appendGroup(groups, attr.Key)
		for _, sub := range attr.Value.Group() {
			h.writeAttr(b, childGroups, sub)
		}
		return
	}

	keyParts := append([]string{}, groups...)
	if attr.Key != "" {
		keyParts = append(keyParts, attr.Key)
	}
	if len(keyParts) == 0 {
		return
	}

	b.WriteByte(' ')
	b.WriteString(strings.Join(keyParts, "."))
	b.WriteByte('=')
	b.WriteString(formatValue(attr.Value))
}

func formatValue(v slog.Value) string {
	switch v.Kind() {
	case slog.KindString:
		return v.String()
	case slog.KindTime:
		return v.Time().Format(time.RFC3339Nano)
	case slog.KindDuration:
		return v.Duration().String()
	case slog.KindFloat64:
		return fmt.Sprintf("%f", v.Float64())
	case slog.KindInt64:
		return fmt.Sprintf("%d", v.Int64())
	case slog.KindUint64:
		return fmt.Sprintf("%d", v.Uint64())
	case slog.KindBool:
		if v.Bool() {
			return "true"
		}
		return "false"
	default:
		return v.String()
	}
}

func appendGroup(groups []string, key string) []string {
	if key == "" {
		return append([]string{}, groups...)
	}
	cp := append([]string{}, groups...)
	return append(cp, key)
}

func resolveLogPath(basePath, service string) string {
	raw := strings.TrimSpace(basePath)
	if raw == "" {
		raw = "./log/app"
	}

	dirHint := strings.HasSuffix(strings.TrimSpace(basePath), "/") || strings.HasSuffix(strings.TrimSpace(basePath), "\\")
	path := filepath.Clean(raw)
	ext := filepath.Ext(path)

	var dir, name string
	switch {
	case dirHint:
		dir = path
		name = "app"
		ext = ".log"
	case ext == "":
		if strings.Contains(path, string(os.PathSeparator)) {
			dir = filepath.Dir(path)
			name = filepath.Base(path)
		} else {
			dir = "."
			name = path
		}
		if name == "" || name == "." {
			name = "app"
		}
		ext = ".log"
	default:
		dir = filepath.Dir(path)
		name = strings.TrimSuffix(filepath.Base(path), ext)
		if name == "" || name == "." {
			name = "app"
		}
	}

	if dir == "" {
		dir = "."
	}
	if service != "" {
		name = fmt.Sprintf("%s-%s", name, service)
	}
	return filepath.Join(dir, name+ext)
}

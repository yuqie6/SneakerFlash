package logger

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	gormlogger "gorm.io/gorm/logger"
)

// NewGormLogger 构建基于 slog 的 GORM 日志器，支持慢查询告警。
func NewGormLogger(level gormlogger.LogLevel, slowThreshold time.Duration) gormlogger.Interface {
	if slowThreshold <= 0 {
		slowThreshold = 500 * time.Millisecond
	}
	return &gormLogger{
		level:         level,
		slowThreshold: slowThreshold,
	}
}

type gormLogger struct {
	level         gormlogger.LogLevel
	slowThreshold time.Duration
}

func (l *gormLogger) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	newLogger := *l
	newLogger.level = level
	return &newLogger
}

func (l *gormLogger) Info(ctx context.Context, msg string, data ...any) {
	if l.level < gormlogger.Info {
		return
	}
	logWithAttrs(ctx, slog.LevelInfo, msg, attrsFromData(data))
}

func (l *gormLogger) Warn(ctx context.Context, msg string, data ...any) {
	if l.level < gormlogger.Warn {
		return
	}
	logWithAttrs(ctx, slog.LevelWarn, msg, attrsFromData(data))
}

func (l *gormLogger) Error(ctx context.Context, msg string, data ...any) {
	if l.level < gormlogger.Error {
		return
	}
	logWithAttrs(ctx, slog.LevelError, msg, attrsFromData(data))
}

func (l *gormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.level == gormlogger.Silent {
		return
	}
	elapsed := time.Since(begin)
	sql, rows := fc()
	attrs := []slog.Attr{
		slog.Duration("elapsed", elapsed),
		slog.String("sql", sql),
	}
	if rows != -1 {
		attrs = append(attrs, slog.Int64("rows", rows))
	}

	switch {
	case err != nil && l.level >= gormlogger.Error:
		attrs = append(attrs, slog.Any("err", err))
		logWithAttrs(ctx, slog.LevelError, "数据库执行失败", attrs)
	case l.slowThreshold > 0 && elapsed > l.slowThreshold && l.level >= gormlogger.Warn:
		attrs = append(attrs, slog.Duration("slow_threshold", l.slowThreshold))
		logWithAttrs(ctx, slog.LevelWarn, "数据库执行缓慢", attrs)
	case l.level >= gormlogger.Info:
		logWithAttrs(ctx, slog.LevelInfo, "数据库执行", attrs)
	}
}

func attrsFromData(data []interface{}) []slog.Attr {
	if len(data) == 0 {
		return nil
	}
	attrs := make([]slog.Attr, 0, len(data))
	for idx, v := range data {
		attrs = append(attrs, slog.Any(fmt.Sprintf("arg_%d", idx), v))
	}
	return attrs
}

func logWithAttrs(ctx context.Context, level slog.Level, msg string, attrs []slog.Attr) {
	args := make([]any, len(attrs))
	for i, attr := range attrs {
		args[i] = attr
	}
	slog.Log(ctx, level, msg, args...)
}

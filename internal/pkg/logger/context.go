package logger

import (
	"context"
	"log/slog"
)

type contextAttrsKey struct{}

// ContextWithAttrs 将结构化属性附着到上下文中，使得后续 slog.*Context 调用自动带出关键信息。
func ContextWithAttrs(ctx context.Context, attrs ...slog.Attr) context.Context {
	if len(attrs) == 0 {
		return ctx
	}
	existing := contextAttrs(ctx)
	merged := make([]slog.Attr, 0, len(existing)+len(attrs))
	merged = append(merged, existing...)
	merged = append(merged, attrs...)
	return context.WithValue(ctx, contextAttrsKey{}, merged)
}

// ContextWithValues 以 key/value 形式附加上下文属性，要求 key 为 string。
func ContextWithValues(ctx context.Context, kv ...any) context.Context {
	if len(kv) == 0 {
		return ctx
	}
	attrs := make([]slog.Attr, 0, len(kv)/2)
	for i := 0; i+1 < len(kv); i += 2 {
		key, ok := kv[i].(string)
		if !ok {
			continue
		}
		attrs = append(attrs, slog.Any(key, kv[i+1]))
	}
	return ContextWithAttrs(ctx, attrs...)
}

// ContextAttrs 返回上下文里缓存的属性切片（只读）。
func ContextAttrs(ctx context.Context) []slog.Attr {
	out := contextAttrs(ctx)
	if len(out) == 0 {
		return nil
	}
	cp := make([]slog.Attr, len(out))
	copy(cp, out)
	return cp
}

func contextAttrs(ctx context.Context) []slog.Attr {
	if ctx == nil {
		return nil
	}
	if attrs, ok := ctx.Value(contextAttrsKey{}).([]slog.Attr); ok {
		return attrs
	}
	return nil
}

package logutils

import (
	"context"
	"log/slog"
)

const (
	LayerHandler    = "handler"
	LayerUsecase    = "usecase"
	LayerRepository = "repository"
)

func logger(ctx context.Context) *slog.Logger {
	rid := RequestIDFromContext(ctx)
	if rid == "" {
		return slog.Default()
	}
	return slog.Default().With(slog.String("request_id", rid))
}

func withBase(layer, op string, attrs []any) []any {
	out := []any{slog.String("layer", layer), slog.String("op", op)}
	return append(out, attrs...)
}

func Debug(ctx context.Context, layer, op, msg string, attrs ...any) {
	logger(ctx).Debug(msg, withBase(layer, op, attrs)...)
}

func Info(ctx context.Context, layer, op, msg string, attrs ...any) {
	logger(ctx).Info(msg, withBase(layer, op, attrs)...)
}

func Warn(ctx context.Context, layer, op, msg string, attrs ...any) {
	logger(ctx).Warn(msg, withBase(layer, op, attrs)...)
}

func Error(ctx context.Context, layer, op, msg string, attrs ...any) {
	logger(ctx).Error(msg, withBase(layer, op, attrs)...)
}

package api

import (
	"context"
	"log/slog"

	logutils "solvi/internal/shared/logUtils"

	"github.com/labstack/echo/v5"
)

func requestCtx(c *echo.Context) context.Context {
	return c.Request().Context()
}

func logHandlerDebug(ctx context.Context, op, msg string, httpStatus int, attrs ...any) {
	attrs = append(attrs, slog.Int("http_status", httpStatus))
	logutils.Debug(ctx, logutils.LayerHandler, op, msg, attrs...)
}

func logHandlerWarn(ctx context.Context, op, msg string, httpStatus int, attrs ...any) {
	attrs = append(attrs, slog.Int("http_status", httpStatus))
	logutils.Warn(ctx, logutils.LayerHandler, op, msg, attrs...)
}

func logHandlerError(ctx context.Context, op, msg string, err error, attrs ...any) {
	if err != nil {
		attrs = append(attrs, slog.String("err", err.Error()))
	}
	logutils.Error(ctx, logutils.LayerHandler, op, msg, attrs...)
}

func logHandlerInfo(ctx context.Context, op, msg string, attrs ...any) {
	logutils.Info(ctx, logutils.LayerHandler, op, msg, attrs...)
}

func handlerAttrs(c *echo.Context) []any {
	req := c.Request()
	return []any{
		slog.String("method", req.Method),
		slog.String("path", req.URL.Path),
	}
}

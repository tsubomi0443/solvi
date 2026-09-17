package page

import (
	"context"
	"log/slog"

	logutils "solvi/internal/shared/logUtils"

	"github.com/labstack/echo/v5"
)

func requestCtx(c *echo.Context) context.Context {
	return c.Request().Context()
}

func logPageWarn(ctx context.Context, op, msg string, attrs ...any) {
	logutils.Warn(ctx, logutils.LayerHandler, op, msg, attrs...)
}

func logPageDebug(ctx context.Context, op, msg string, attrs ...any) {
	logutils.Debug(ctx, logutils.LayerHandler, op, msg, attrs...)
}

func logPageError(ctx context.Context, op, msg string, err error, attrs ...any) {
	if err != nil {
		attrs = append(attrs, slog.String("err", err.Error()))
	}
	logutils.Error(ctx, logutils.LayerHandler, op, msg, attrs...)
}

func pageAttrs(c *echo.Context) []any {
	req := c.Request()
	return []any{
		slog.String("method", req.Method),
		slog.String("path", req.URL.Path),
	}
}

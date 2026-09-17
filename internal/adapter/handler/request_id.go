package handler

import (
	logutils "solvi/internal/shared/logUtils"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
)

const requestIDHeader = "X-Request-ID"

func RequestID() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			req := c.Request()
			id := req.Header.Get(requestIDHeader)
			if id == "" {
				id = uuid.NewString()
			}
			c.Response().Header().Set(requestIDHeader, id)
			ctx := logutils.WithRequestID(req.Context(), id)
			c.SetRequest(req.WithContext(ctx))
			return next(c)
		}
	}
}

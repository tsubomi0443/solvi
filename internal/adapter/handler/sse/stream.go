package sse

import (
	"fmt"
	"net/http"

	"solvi/internal/adapter/handler/authctx"

	"github.com/labstack/echo/v5"
)

type Handler struct {
	hub *Hub
}

func NewHandler(hub *Hub) *Handler {
	return &Handler{hub: hub}
}

func (h *Handler) Stream(c *echo.Context) error {
	claims := authctx.Claims(c)
	if claims == nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}
	w := c.Response()
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)

	client := &Client{
		UserID:      claims.UserID,
		IsSupporter: claims.IsSupporter,
		Send:        make(chan Event, 8),
	}
	h.hub.Register(client)
	defer h.hub.Unregister(client)

	ctx := c.Request().Context()
	flusher, ok := w.(http.Flusher)
	if !ok {
		return echo.NewHTTPError(http.StatusInternalServerError, "streaming unsupported")
	}
	for {
		select {
		case <-ctx.Done():
			return nil
		case ev := <-client.Send:
			if _, err := fmt.Fprintf(w, "event: %s\ndata: %s\n\n", ev.Event, ev.Data); err != nil {
				return nil
			}
			flusher.Flush()
		}
	}
}

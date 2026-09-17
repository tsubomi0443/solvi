package api

import (
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v5"
)

func (h *Handler) ListTags(c *echo.Context) error {
	const op = "api.ListTags"
	ctx := requestCtx(c)
	tags, err := h.deps.Tag.List(ctx)
	if err != nil {
		logHandlerDebug(ctx, op, "タグ一覧取得失敗", http.StatusInternalServerError, handlerAttrs(c)...)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, tags)
}

func (h *Handler) RenameTag(c *echo.Context) error {
	const op = "api.RenameTag"
	ctx := requestCtx(c)
	var req struct {
		From string `json:"from"`
		To   string `json:"to"`
	}
	if err := c.Bind(&req); err != nil {
		logHandlerWarn(ctx, op, "リクエスト不正", http.StatusBadRequest, handlerAttrs(c)...)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}
	if err := h.deps.Tag.Rename(ctx, req.From, req.To); err != nil {
		logHandlerDebug(ctx, op, "タグ名変更失敗", http.StatusBadRequest, append(handlerAttrs(c), slog.String("from", req.From), slog.String("to", req.To))...)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{"ok": "true"})
}

func (h *Handler) DeleteTag(c *echo.Context) error {
	const op = "api.DeleteTag"
	ctx := requestCtx(c)
	var req struct {
		Name string `json:"name"`
	}
	if err := c.Bind(&req); err != nil {
		logHandlerWarn(ctx, op, "リクエスト不正", http.StatusBadRequest, handlerAttrs(c)...)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}
	if err := h.deps.Tag.Delete(ctx, req.Name); err != nil {
		logHandlerDebug(ctx, op, "タグ削除失敗", http.StatusBadRequest, append(handlerAttrs(c), slog.String("name", req.Name))...)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{"ok": "true"})
}

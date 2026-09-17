package api

import (
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v5"
)

func (h *Handler) GetUserIcon(c *echo.Context) error {
	const op = "api.GetUserIcon"
	ctx := requestCtx(c)
	uuid := c.Param("uuid")
	userProfile, err := h.deps.Setting.GetProfileByUUID(ctx, uuid)
	if err != nil {
		logHandlerDebug(ctx, op, "ユーザ取得失敗", http.StatusForbidden, append(handlerAttrs(c), slog.String("user_uuid", uuid))...)
		return c.JSON(http.StatusForbidden, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[any]any{"icon": userProfile.IconBase64})
}

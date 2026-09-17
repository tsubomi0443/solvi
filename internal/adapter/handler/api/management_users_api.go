package api

import (
	"log/slog"
	"net/http"

	"solvi/internal/adapter/handler/authctx"
	mnguc "solvi/internal/application/usecase/management_usecase"

	"github.com/labstack/echo/v5"
)

func (h *Handler) ListUsers(c *echo.Context) error {
	const op = "api.ListUsers"
	ctx := requestCtx(c)
	users, err := h.deps.Management.ListUsers(ctx)
	if err != nil {
		logHandlerDebug(ctx, op, "ユーザ一覧取得失敗", http.StatusInternalServerError, handlerAttrs(c)...)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, users)
}

func (h *Handler) UpdateUser(c *echo.Context) error {
	const op = "api.UpdateUser"
	ctx := requestCtx(c)
	claims := authctx.Claims(c)
	var req struct {
		IsSupporter *bool `json:"isSupporter"`
		IsAdmin     *bool `json:"isAdmin"`
	}
	if err := c.Bind(&req); err != nil {
		logHandlerWarn(ctx, op, "リクエスト不正", http.StatusBadRequest, handlerAttrs(c)...)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}
	targetUUID := c.Param("uuid")
	if err := h.deps.Management.UpdateUser(ctx, targetUUID, claims.UUID, mnguc.UserUpdateInput{
		IsSupporter: req.IsSupporter,
		IsAdmin:     req.IsAdmin,
	}); err != nil {
		logHandlerDebug(ctx, op, "ユーザ更新失敗", http.StatusBadRequest, append(handlerAttrs(c), slog.String("target_uuid", targetUUID))...)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{"ok": "true"})
}

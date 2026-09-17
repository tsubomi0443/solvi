package api

import (
	"log/slog"
	"net/http"

	"solvi/internal/adapter/handler/authctx"

	"github.com/labstack/echo/v5"
)

func (h *Handler) UpdateSetting(c *echo.Context) error {
	const op = "api.UpdateSetting"
	ctx := requestCtx(c)
	claims := authctx.Claims(c)
	var req struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}
	if err := c.Bind(&req); err != nil {
		logHandlerWarn(ctx, op, "リクエスト不正", http.StatusBadRequest, handlerAttrs(c)...)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}
	user, err := h.deps.Setting.UpdateProfile(ctx, claims.UserID, req.Name, req.Email)
	if err != nil {
		logHandlerDebug(ctx, op, "プロフィール更新失敗", http.StatusInternalServerError, append(handlerAttrs(c), slog.Uint64("user_id", uint64(claims.UserID)))...)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, user)
}

func (h *Handler) GetIcon(c *echo.Context) error {
	return c.JSON(http.StatusOK, nil)
}

func (h *Handler) UploadIcon(c *echo.Context) error {
	const op = "api.UploadIcon"
	ctx := requestCtx(c)
	claims := authctx.Claims(c)
	file, err := c.FormFile("icon")
	if err != nil {
		logHandlerWarn(ctx, op, "ファイル未指定", http.StatusBadRequest, handlerAttrs(c)...)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "file required"})
	}
	src, err := file.Open()
	if err != nil {
		logHandlerError(ctx, op, "ファイル開く失敗", err, append(handlerAttrs(c), slog.Uint64("user_id", uint64(claims.UserID)))...)
		return err
	}
	defer src.Close()
	if err := h.deps.Setting.UploadIcon(ctx, claims.UserID, file.Filename, src); err != nil {
		logHandlerDebug(ctx, op, "アイコンアップロード失敗", http.StatusInternalServerError, append(handlerAttrs(c), slog.Uint64("user_id", uint64(claims.UserID)))...)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{"ok": "true"})
}

func (h *Handler) DeleteIcon(c *echo.Context) error {
	const op = "api.DeleteIcon"
	ctx := requestCtx(c)
	claims := authctx.Claims(c)
	if err := h.deps.Setting.DeleteIcon(ctx, claims.UserID); err != nil {
		logHandlerDebug(ctx, op, "アイコン削除失敗", http.StatusInternalServerError, append(handlerAttrs(c), slog.Uint64("user_id", uint64(claims.UserID)))...)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{"ok": "true"})
}

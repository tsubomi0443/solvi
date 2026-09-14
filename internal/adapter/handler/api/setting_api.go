package api

import (
	"net/http"

	"solvi/internal/adapter/handler/authctx"

	"github.com/labstack/echo/v5"
)

func (h *Handler) UpdateSetting(c *echo.Context) error {
	claims := authctx.Claims(c)
	var req struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}
	user, err := h.deps.Setting.UpdateProfile(claims.UserID, req.Name, req.Email)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, user)
}

func (h *Handler) UploadIcon(c *echo.Context) error {
	claims := authctx.Claims(c)
	file, err := c.FormFile("icon")
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "file required"})
	}
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()
	if err := h.deps.Setting.UploadIcon(claims.UserID, file.Filename, src); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{"ok": "true"})
}

func (h *Handler) DeleteIcon(c *echo.Context) error {
	claims := authctx.Claims(c)
	if err := h.deps.Setting.DeleteIcon(claims.UserID); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{"ok": "true"})
}

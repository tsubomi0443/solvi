package api

import (
	"net/http"

	"solvi/internal/adapter/handler/authctx"
	mnguc "solvi/internal/application/usecase/management_usecase"

	"github.com/labstack/echo/v5"
)

func (h *Handler) ListUsers(c *echo.Context) error {
	users, err := h.deps.Management.ListUsers()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, users)
}

func (h *Handler) UpdateUser(c *echo.Context) error {
	claims := authctx.Claims(c)
	var req struct {
		IsSupporter *bool `json:"isSupporter"`
		IsAdmin     *bool `json:"isAdmin"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}
	if err := h.deps.Management.UpdateUser(c.Param("uuid"), claims.UUID, mnguc.UserUpdateInput{
		IsSupporter: req.IsSupporter,
		IsAdmin:     req.IsAdmin,
	}); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{"ok": "true"})
}

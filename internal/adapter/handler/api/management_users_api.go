package api

import (
	"net/http"

	"github.com/labstack/echo/v5"
)

func (h *Handler) ListUsers(c *echo.Context) error {
	users, err := h.deps.Management.ListUsers()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, users)
}

func (h *Handler) UpdateIsSupporter(c *echo.Context) error {
	var req struct {
		IsSupporter bool `json:"isSupporter"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}
	if err := h.deps.Management.UpdateIsSupporter(c.Param("uuid"), req.IsSupporter); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{"ok": "true"})
}

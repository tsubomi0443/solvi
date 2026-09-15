package api

import (
	"net/http"

	"github.com/labstack/echo/v5"
)

func (h *Handler) GetUserIcon(c *echo.Context) error {
	uuid := c.Param("uuid")
	userProfile, err := h.deps.Setting.GetProfileByUUID(uuid)
	if err != nil {
		return c.JSON(http.StatusForbidden, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[any]any{"icon": userProfile.IconBase64})
}

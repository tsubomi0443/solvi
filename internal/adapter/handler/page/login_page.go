package page

import (
	"net/http"

	"github.com/labstack/echo/v5"
)

func (h *Handler) LoginPage(c *echo.Context) error {
	return c.Render(http.StatusOK, "login.html", map[string]interface{}{})
}

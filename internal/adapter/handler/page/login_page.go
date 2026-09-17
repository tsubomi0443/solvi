package page

import (
	"html/template"
	"net/http"

	"github.com/labstack/echo/v5"
)

func (h *Handler) LoginPage(c *echo.Context) error {
	data := map[string]template.HTML{"Logo": template.HTML(h.logo())}
	return c.Render(http.StatusOK, "login.html", data)
}

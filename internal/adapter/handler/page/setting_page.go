package page

import (
	"net/http"

	"github.com/labstack/echo/v5"
)

func (h *Handler) SettingPage(c *echo.Context) error {
	data := h.baseData(c, "setting")
	data["UserJSON"] = mustJSON(data["User"])
	return c.Render(http.StatusOK, "setting.html", data)
}

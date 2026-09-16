package page

import (
	"net/http"

	"github.com/labstack/echo/v5"
)

func (h *Handler) ManagementLogsPage(c *echo.Context) error {
	data := h.baseData(c, "management-logs")
	dates := []string{}
	if h.deps.Log != nil {
		dates = h.deps.Log.ListAvailableDates()
	}
	data["LogDatesJSON"] = mustJSON(dates)
	return c.Render(http.StatusOK, "management_logs.html", data)
}

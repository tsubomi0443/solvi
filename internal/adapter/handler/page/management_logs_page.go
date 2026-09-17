package page

import (
	"net/http"

	"github.com/labstack/echo/v5"
)

func (h *Handler) ManagementLogsPage(c *echo.Context) error {
	const op = "page.ManagementLogsPage"
	ctx := requestCtx(c)
	data := h.baseData(c, "management-logs")
	dates := []string{}
	if h.deps.Log != nil {
		dates = h.deps.Log.ListAvailableDates(ctx)
	}
	data["LogDatesJSON"] = mustJSON(c, dates)
	logPageDebug(ctx, op, "ページ描画", pageAttrs(c)...)
	return c.Render(http.StatusOK, "management_logs.html", data)
}

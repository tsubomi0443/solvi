package page

import (
	"net/http"

	"github.com/labstack/echo/v5"
)

func (h *Handler) SettingPage(c *echo.Context) error {
	const op = "page.SettingPage"
	ctx := requestCtx(c)
	data := h.baseData(c, "setting")
	data["UserJSON"] = mustJSON(c, data["User"])
	logPageDebug(ctx, op, "ページ描画", pageAttrs(c)...)
	return c.Render(http.StatusOK, "setting.html", data)
}

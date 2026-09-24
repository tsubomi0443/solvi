package page

import (
	"net/http"

	"github.com/labstack/echo/v5"
)

func (h *Handler) FAQPage(c *echo.Context) error {
	const op = "page.FAQPage"
	ctx := requestCtx(c)
	items, err := h.deps.Question.ListSummaries(ctx)
	if err != nil {
		logPageError(ctx, op, "FAQ一覧取得失敗", err, pageAttrs(c)...)
		return err
	}
	data := h.baseData(c, "faq")
	data["Summaries"] = items
	data["SummariesJSON"] = mustJSON(c, items)
	logPageDebug(ctx, op, "ページ描画", pageAttrs(c)...)
	return c.Render(http.StatusOK, "faq.html", data)
}

package page

import (
	"net/http"

	"github.com/labstack/echo/v5"
)

func (h *Handler) QuestionNewPage(c *echo.Context) error {
	const op = "page.QuestionNewPage"
	ctx := requestCtx(c)
	data := h.baseData(c, "new")
	data["isHumanSupportRequired"] = true
	logPageDebug(ctx, op, "ページ描画", pageAttrs(c)...)
	return c.Render(http.StatusOK, "question_new.html", data)
}

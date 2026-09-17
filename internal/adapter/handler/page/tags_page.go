package page

import (
	"net/http"

	"github.com/labstack/echo/v5"
)

func (h *Handler) TagsPage(c *echo.Context) error {
	const op = "page.TagsPage"
	ctx := requestCtx(c)
	tags, err := h.deps.Tag.List(ctx)
	if err != nil {
		logPageError(ctx, op, "タグ一覧取得失敗", err, pageAttrs(c)...)
		return err
	}
	data := h.baseData(c, "tags")
	data["Tags"] = tags
	data["TagsJSON"] = mustJSON(c, tags)
	logPageDebug(ctx, op, "ページ描画", pageAttrs(c)...)
	return c.Render(http.StatusOK, "tags.html", data)
}

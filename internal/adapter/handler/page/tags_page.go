package page

import (
	"net/http"

	"github.com/labstack/echo/v5"
)

func (h *Handler) TagsPage(c *echo.Context) error {
	tags, err := h.deps.Tag.List()
	if err != nil {
		return err
	}
	data := h.baseData(c, "tags")
	data["Tags"] = tags
	data["TagsJSON"] = mustJSON(tags)
	return c.Render(http.StatusOK, "tags.html", data)
}

package page

import (
	"net/http"

	"github.com/labstack/echo/v5"
)

func (h *Handler) QuestionNewPage(c *echo.Context) error {
	data := h.baseData(c, "new")
	return c.Render(http.StatusOK, "question_new.html", data)
}

package page

import (
	"net/http"

	"solvi/internal/adapter/handler/authctx"

	"github.com/labstack/echo/v5"
)

func (h *Handler) QuestionDetailPage(c *echo.Context) error {
	claims := authctx.Claims(c)
	uuid := c.Param("uuid")
	detail, err := h.deps.Question.Get(claims.UserID, claims.IsSupporter, claims.IsAdmin, uuid)
	if err != nil {
		return c.Redirect(http.StatusFound, "/")
	}
	data := h.baseData(c, "detail")
	data["Question"] = detail
	data["QuestionJSON"] = mustJSON(detail)
	data["UserJSON"] = mustJSON(data["User"])
	return c.Render(http.StatusOK, "question_detail.html", data)
}

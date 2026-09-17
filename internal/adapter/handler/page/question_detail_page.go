package page

import (
	"log/slog"
	"net/http"

	"solvi/internal/adapter/handler/authctx"

	"github.com/labstack/echo/v5"
)

func (h *Handler) QuestionDetailPage(c *echo.Context) error {
	const op = "page.QuestionDetailPage"
	ctx := requestCtx(c)
	claims := authctx.Claims(c)
	uuid := c.Param("uuid")
	detail, err := h.deps.Question.Get(ctx, claims.UserID, claims.IsSupporter, claims.IsAdmin, uuid)
	if err != nil {
		logPageWarn(ctx, op, "質問取得失敗、一覧へリダイレクト", append(pageAttrs(c), slog.String("question_uuid", uuid), slog.String("err", err.Error()))...)
		return c.Redirect(http.StatusFound, "/")
	}
	data := h.baseData(c, "detail")
	data["Question"] = detail
	data["QuestionJSON"] = mustJSON(c, detail)
	data["UserJSON"] = mustJSON(c, data["User"])
	logPageDebug(ctx, op, "ページ描画", append(pageAttrs(c), slog.String("question_uuid", uuid))...)
	return c.Render(http.StatusOK, "question_detail.html", data)
}

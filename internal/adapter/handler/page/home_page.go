package page

import (
	"log/slog"
	"net/http"

	"solvi/internal/adapter/handler/authctx"

	"github.com/labstack/echo/v5"
)

func (h *Handler) HomePage(c *echo.Context) error {
	const op = "page.HomePage"
	ctx := requestCtx(c)
	claims := authctx.Claims(c)
	items, err := h.deps.Question.List(ctx, claims.UserID, claims.IsSupporter, claims.IsAdmin)
	if err != nil {
		logPageError(ctx, op, "質問一覧取得失敗", err, append(pageAttrs(c), slog.Uint64("user_id", uint64(claims.UserID)))...)
		return err
	}
	data := h.baseData(c, "home")
	data["Questions"] = items
	data["QuestionsJSON"] = mustJSON(c, items)
	logPageDebug(ctx, op, "ページ描画", pageAttrs(c)...)
	return c.Render(http.StatusOK, "home.html", data)
}

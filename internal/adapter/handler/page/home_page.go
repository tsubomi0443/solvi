package page

import (
	"log/slog"
	"net/http"

	"solvi/internal/adapter/handler/authctx"

	"github.com/labstack/echo/v5"
)

func (h *Handler) HomePage(c *echo.Context) error {
	claims := authctx.Claims(c)
	items, err := h.deps.Question.List(claims.UserID, claims.IsSupporter)
	if err != nil {
		return err
	}
	data := h.baseData(c, "home")
	data["Questions"] = items
	data["QuestionsJSON"] = mustJSON(items)
	slog.Info("home-data", slog.Any("data", data))
	return c.Render(http.StatusOK, "home.html", data)
}

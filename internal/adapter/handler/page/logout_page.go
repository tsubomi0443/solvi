package page

import (
	"net/http"

	"solvi/internal/adapter/handler/authctx"

	"github.com/labstack/echo/v5"
)

func (h *Handler) LogoutPage(c *echo.Context) error {
	c.SetCookie(&http.Cookie{Name: authctx.CookieNameToken, Value: "", Path: "/", MaxAge: -1})
	return c.Render(http.StatusOK, "logout.html", map[string]interface{}{})
}

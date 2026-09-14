package api

import (
	"net/http"

	"solvi/internal/adapter/handler/authctx"
	outputmodel "solvi/internal/application/model/output_model"

	"github.com/labstack/echo/v5"
)

func (h *Handler) Login(c *echo.Context) error {
	var req struct {
		Method   string `json:"method"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}
	var token string
	var user outputmodel.UserOutput
	var err error
	switch req.Method {
	case "ldap":
		token, user, err = h.deps.Auth.LoginLDAP(req.Email, req.Password)
	case "basic":
		token, user, err = h.deps.Auth.LoginBasic(req.Email, req.Password)
	default:
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid method"})
	}
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": err.Error()})
	}
	setTokenCookie(c, token)
	return c.JSON(http.StatusOK, user)
}

func setTokenCookie(c *echo.Context, token string) {
	c.SetCookie(&http.Cookie{Name: authctx.CookieNameToken, Value: token, Path: "/", HttpOnly: true, SameSite: http.SameSiteLaxMode, MaxAge: 86400})
}

func (h *Handler) Logout(c *echo.Context) error {
	c.SetCookie(&http.Cookie{Name: authctx.CookieNameToken, Value: "", Path: "/", MaxAge: -1})
	return c.JSON(http.StatusOK, map[string]string{"ok": "true"})
}

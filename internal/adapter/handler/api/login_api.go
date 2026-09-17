package api

import (
	"log/slog"
	"net/http"

	"solvi/internal/adapter/handler/authctx"
	outputmodel "solvi/internal/application/model/output_model"

	"github.com/labstack/echo/v5"
)

func (h *Handler) Login(c *echo.Context) error {
	const op = "api.Login"
	ctx := requestCtx(c)
	var req struct {
		Method   string `json:"method"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.Bind(&req); err != nil {
		logHandlerWarn(ctx, op, "リクエスト不正", http.StatusBadRequest, handlerAttrs(c)...)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}
	var (
		token string
		user  outputmodel.UserOutput
		err   error
	)
	switch req.Method {
	case "ldap":
		token, user, err = h.deps.Auth.LoginLDAP(ctx, req.Email, req.Password)
	case "basic":
		token, user, err = h.deps.Auth.LoginBasic(ctx, req.Email, req.Password)
	default:
		logHandlerWarn(ctx, op, "認証方式不正", http.StatusBadRequest, append(handlerAttrs(c), slog.String("method", req.Method))...)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid method"})
	}

	if err != nil {
		logHandlerDebug(ctx, op, "ログイン失敗", http.StatusUnauthorized, append(handlerAttrs(c), slog.String("email", req.Email), slog.String("auth_method", req.Method))...)
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": err.Error()})
	}
	setTokenCookie(c, token)
	return c.JSON(http.StatusOK, user)
}

func setTokenCookie(c *echo.Context, token string) {
	c.SetCookie(&http.Cookie{Name: authctx.CookieNameToken, Value: token, Path: "/", HttpOnly: true, SameSite: http.SameSiteLaxMode, MaxAge: 86400})
}

func (h *Handler) Logout(c *echo.Context) error {
	const op = "api.Logout"
	ctx := requestCtx(c)
	c.SetCookie(&http.Cookie{Name: authctx.CookieNameToken, Value: "", Path: "/", MaxAge: -1})
	logHandlerInfo(ctx, op, "ログアウト成功", handlerAttrs(c)...)
	return c.JSON(http.StatusOK, map[string]string{"ok": "true"})
}

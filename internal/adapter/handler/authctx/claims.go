package authctx

import (
	"solvi/internal/shared/auth"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v5"
)

const CookieNameToken = "access_token"

func Claims(c *echo.Context) *auth.CustomClaims {
	claims, ok := c.Get("user").(*jwt.Token)
	if !ok || claims == nil {
		return nil
	}
	custom, ok := claims.Claims.(*auth.CustomClaims)
	if !ok {
		return nil
	}
	return custom
}

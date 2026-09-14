package handler

import (
	"net/http"

	"solvi/internal/adapter/handler/authctx"
	"solvi/internal/shared/auth"
	"solvi/internal/shared/config"

	"github.com/golang-jwt/jwt/v5"
	echojwt "github.com/labstack/echo-jwt/v5"
	"github.com/labstack/echo/v5"
)

func JWTConfig() echo.MiddlewareFunc {
	return echojwt.WithConfig(echojwt.Config{
		NewClaimsFunc: func(c *echo.Context) jwt.Claims { return new(auth.CustomClaims) },
		SigningKey:    []byte(config.GetJWTKey()),
		TokenLookup:   "cookie:" + authctx.CookieNameToken,
		ErrorHandler: func(c *echo.Context, err error) error {
			return c.Redirect(http.StatusFound, "/login")
		},
	})
}

func APIJWTConfig() echo.MiddlewareFunc {
	return echojwt.WithConfig(echojwt.Config{
		NewClaimsFunc: func(c *echo.Context) jwt.Claims { return new(auth.CustomClaims) },
		SigningKey:    []byte(config.GetJWTKey()),
		TokenLookup:   "cookie:" + authctx.CookieNameToken,
		ErrorHandler: func(c *echo.Context, err error) error {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		},
	})
}

func SupporterOnlyPage(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c *echo.Context) error {
		claims := authctx.Claims(c)
		if claims == nil || !claims.IsSupporter {
			return c.Redirect(http.StatusFound, "/")
		}
		return next(c)
	}
}

func AdminOnly(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c *echo.Context) error {
		claims := authctx.Claims(c)
		if claims == nil || !claims.IsAdmin {
			return c.Redirect(http.StatusFound, "/")
		}
		return next(c)
	}
}

func SupporterOnly(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c *echo.Context) error {
		claims := authctx.Claims(c)
		if claims == nil || !claims.IsSupporter {
			return c.JSON(http.StatusForbidden, map[string]string{"error": "forbidden"})
		}
		return next(c)
	}
}

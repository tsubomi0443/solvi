package auth

import (
	"github.com/golang-jwt/jwt/v5"
)

type CustomClaims struct {
	UserID      uint   `json:"user_id"`
	UUID        string `json:"uuid"`
	Email       string `json:"email"`
	Name        string `json:"name"`
	IsSupporter bool   `json:"is_supporter"`
	IsAdmin     bool   `json:"is_admin"`
	jwt.RegisteredClaims
}

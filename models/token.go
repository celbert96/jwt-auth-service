package models

import (
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type ClientReadableToken struct {
	ExpiresAt int64   `json:"expires_at"`
	UserRoles []Roles `json:"roles"`
}
type TokenClaims struct {
	jwt.RegisteredClaims
	UserRoles []Roles `json:"roles"`
}

func MintToken(userid int, userroles []Roles, expires time.Time) (string, error) {
	claims := TokenClaims{
		jwt.RegisteredClaims{
			Issuer:    "jwt-auth-service",
			Subject:   strconv.Itoa(userid),
			ExpiresAt: jwt.NewNumericDate(expires),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		userroles,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"iss":   claims.Issuer,
		"sub":   claims.Subject,
		"exp":   claims.ExpiresAt,
		"iat":   claims.IssuedAt,
		"roles": claims.UserRoles,
	})

	return token.SignedString([]byte(os.Getenv("JWT_AUTH_SERVICE_SECRET_KEY")))

}

func ValidateToken(tokenStr string) (*jwt.Token, TokenClaims, error) {
	claims := TokenClaims{}
	token, err := jwt.ParseWithClaims(tokenStr, &claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_AUTH_SERVICE_SECRET_KEY")), nil
	})

	return token, claims, err
}

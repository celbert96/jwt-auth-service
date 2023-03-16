package utils

import (
	"fmt"
	"log"
	"strings"

	"github.com/gin-gonic/gin"
)

func GetAuthCookiesFromContext(c *gin.Context) (string, string, error) {
	/* Get auth token */
	authTokenCookie, err := c.Request.Cookie("authtoken")
	if err != nil {
		log.Println(err.Error())
		return "", "", err
	}

	/* Get session token */
	refreshTokenCookie, err := c.Request.Cookie("refreshtoken")
	if err != nil {
		log.Println(err.Error())
		return "", "", err
	}

	return authTokenCookie.Value, refreshTokenCookie.Value, nil
}

func GetBearerTokenFromContext(c *gin.Context) (string, error) {
	authHeader := c.Request.Header.Get("Authorization")
	if authHeader == "" {
		return "", fmt.Errorf("missing authorization header")
	}

	authToken := strings.TrimPrefix(authHeader, "Bearer ")
	if authToken == authHeader {
		return "", fmt.Errorf("malformed authorization header")
	}

	return authToken, nil
}

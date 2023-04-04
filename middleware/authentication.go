package middleware

import (
	"jwt-auth-service/models"
	"jwt-auth-service/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

func CookieTokenAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authTokenStr, _, err := utils.GetAuthCookiesFromContext(c)
		if err != nil {
			c.IndentedJSON(http.StatusForbidden, models.ErrorResponse{ErrorMessage: "Could not parse auth token(s)"})
			c.Abort()
			return
		}

		_, _, err = models.ValidateToken(authTokenStr)

		if err != nil {
			c.IndentedJSON(http.StatusForbidden, models.ErrResponseForHttpStatus(http.StatusForbidden))
			c.Abort()
			return
		}

		c.Next()
	}
}

func BearerTokenAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authTokenStr, err := utils.GetBearerTokenFromContext(c)
		if err != nil {
			c.IndentedJSON(http.StatusForbidden, models.ErrorResponse{ErrorMessage: err.Error()})
			c.Abort()
			return
		}

		_, _, err = models.ValidateToken(authTokenStr)

		if err != nil {
			c.IndentedJSON(http.StatusForbidden, models.ErrResponseForHttpStatus(http.StatusForbidden))
			c.Abort()
			return
		}

		c.Next()
	}
}

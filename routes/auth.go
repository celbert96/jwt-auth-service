package routes

import (
	"fmt"
	"jwt-auth-service/controllers"
	"jwt-auth-service/models"
	"jwt-auth-service/repositories"
	"jwt-auth-service/utils"
	"log"
	"net/http"
	"net/mail"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type loginresponse struct {
	AuthToken        string                     `json:"auth_token"`
	AuthTokenDetails models.ClientReadableToken `json:"auth_token_details"`
}

type loginrequestbody struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func AddAuthRoutes(rg *gin.RouterGroup) {
	authGroup := rg.Group("/auth")

	authGroup.POST("/login", login)
	authGroup.POST("/register", register)
	authGroup.POST("/refreshtoken", refreshAuthToken)
}

// auth/login
func login(c *gin.Context) {
	var requestBody loginrequestbody

	err := c.BindJSON(&requestBody)
	if err != nil {
		log.Printf("routes > auth.go > login > invalid request > could not parse body")
		c.IndentedJSON(http.StatusBadRequest, models.ErrResponseForHttpStatus(http.StatusBadRequest))
		return
	}
	if errors := requestBody.validate(); errors != nil {
		log.Printf("routes > auth.go > login > invalid request > validation failed")
		c.IndentedJSON(http.StatusBadRequest, models.ErrorResponse{ErrorMessage: "Validation errors occurred", Errors: errors})
		return
	}

	env, ok := c.MustGet("env").(models.Env)
	if !ok {
		log.Println("routes > auth.go > login > env not accessible")
		c.IndentedJSON(http.StatusInternalServerError, models.ErrResponseForHttpStatus(http.StatusInternalServerError))
		return
	}

	repo := repositories.UserRepository{DBConn: env.DB}
	controller := controllers.UserController{UserRepository: repo}

	user, err := controller.GetUserWithCredentials(requestBody.Email, requestBody.Password)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, models.ErrorResponse{ErrorMessage: err.Error()})
		return
	}

	authTokenExpiration := time.Now().Add(time.Minute * 30)
	authTokenString, err := models.MintToken(user.ID, user.UserRoles, authTokenExpiration)
	if err != nil {
		log.Printf("routes > auth.go > getAuthToken > failed to mint auth token")
		c.IndentedJSON(http.StatusInternalServerError, models.ErrorResponse{ErrorMessage: "Could not mint token"})
		return
	}

	refreshTokenExpiration := time.Now().Add(time.Hour * 168) // 1 week
	refreshTokenString, err := models.MintToken(user.ID, user.UserRoles, refreshTokenExpiration)
	err = repo.UpdateRefreshToken(user.ID, refreshTokenString)
	if err != nil {
		log.Printf("routes > auth.go > register > failed to update refresh token")
		c.IndentedJSON(http.StatusInternalServerError, models.ErrResponseForHttpStatus(http.StatusInternalServerError))
		return
	}

	authTokenDetails := models.ClientReadableToken{
		ExpiresAt: authTokenExpiration.Unix(),
		UserRoles: user.UserRoles,
	}

	c.IndentedJSON(http.StatusOK, loginresponse{
		AuthToken:        authTokenString,
		AuthTokenDetails: authTokenDetails,
	})
}

// auth/register
func register(c *gin.Context) {
	var user models.User

	if err := c.BindJSON(&user); err != nil {
		c.IndentedJSON(http.StatusBadRequest, models.ErrResponseForHttpStatus(http.StatusBadRequest))
		return
	}

	env, ok := c.MustGet("env").(models.Env)
	if !ok {
		log.Println("routes > auth.go > register > env not accessible")
		c.IndentedJSON(http.StatusInternalServerError, models.ErrResponseForHttpStatus(http.StatusInternalServerError))
		return
	}

	repo := repositories.UserRepository{DBConn: env.DB}
	controller := controllers.UserController{UserRepository: repo}

	addedUser, errResp := controller.AddUser(user)
	if errResp.ErrorMessage != "" {
		c.IndentedJSON(http.StatusBadRequest, errResp)
		return
	}

	authTokenExpiration := time.Now().Add(time.Minute * 30)
	authTokenString, err := models.MintToken(addedUser.ID, addedUser.UserRoles, authTokenExpiration)
	if err != nil {
		log.Printf("routes > auth.go > register > failed to mint auth token")
		c.IndentedJSON(http.StatusInternalServerError, models.ErrorResponse{ErrorMessage: "Could not mint token"})
		return
	}

	refreshTokenExpiration := time.Now().Add(time.Hour * 168) // 1 week
	refreshTokenString, err := models.MintToken(addedUser.ID, addedUser.UserRoles, refreshTokenExpiration)
	if err != nil {
		log.Printf("routes > auth.go > register > failed to mint refresh token")
		c.IndentedJSON(http.StatusInternalServerError, models.ErrorResponse{ErrorMessage: "Could not mint token"})
		return
	}

	err = repo.UpdateRefreshToken(addedUser.ID, refreshTokenString)
	if err != nil {
		log.Printf("routes > auth.go > register > failed to update refresh token")
		c.IndentedJSON(http.StatusInternalServerError, models.ErrResponseForHttpStatus(http.StatusInternalServerError))
		return
	}

	authTokenDetails := models.ClientReadableToken{
		ExpiresAt: authTokenExpiration.Unix(),
		UserRoles: addedUser.UserRoles,
	}

	c.IndentedJSON(http.StatusOK, loginresponse{
		AuthToken:        authTokenString,
		AuthTokenDetails: authTokenDetails,
	})
}

func refreshAuthToken(c *gin.Context) {
	env, ok := c.MustGet("env").(models.Env)
	if !ok {
		log.Println("routes > auth.go > refreshAuthToken > env not accessible")
		c.IndentedJSON(http.StatusInternalServerError, models.ErrResponseForHttpStatus(http.StatusInternalServerError))
		return
	}

	currentAuthTokenString, err := utils.GetBearerTokenFromContext(c)
	if err != nil {
		c.IndentedJSON(http.StatusUnauthorized, models.ErrResponseForHttpStatus(http.StatusUnauthorized))
		return
	}

	currentAuthToken, claims, _ := models.ValidateToken(currentAuthTokenString)

	if currentAuthToken == nil {
		log.Printf("err: %s", err.Error())
		c.IndentedJSON(http.StatusUnauthorized, models.ErrResponseForHttpStatus(http.StatusUnauthorized))
		return
	}

	userID, err := strconv.Atoi(claims.Subject)
	if err != nil {
		log.Printf("routes > auth.go > refreshAuthToken > invalid user ID %s\n", claims.Subject)
		c.IndentedJSON(http.StatusUnauthorized, models.ErrResponseForHttpStatus(http.StatusUnauthorized))
		return
	}

	repo := repositories.UserRepository{DBConn: env.DB}

	refreshTokenStr, err := repo.GetRefreshToken(userID)
	if err != nil {
		log.Printf("routes > auth.go > refreshAuthToken > could not get refresh token")
		c.IndentedJSON(http.StatusInternalServerError, models.ErrResponseForHttpStatus(http.StatusInternalServerError))
		return
	}

	_, _, err = models.ValidateToken(refreshTokenStr)
	if err != nil {
		log.Printf("routes > auth.go > refreshAuthToken > invalid refresh token > reauthentication needed")
		c.IndentedJSON(http.StatusUnauthorized, models.ErrResponseForHttpStatus(http.StatusUnauthorized))
		return
	}

	newAuthTokenExpiration := time.Now().Add(time.Minute * 30)
	newAuthToken, err := models.MintToken(userID, claims.UserRoles, newAuthTokenExpiration)
	if err != nil {
		log.Printf("routes > auth.go > refreshAuthToken > could not mint new token")
		c.IndentedJSON(http.StatusInternalServerError, models.ErrResponseForHttpStatus(http.StatusInternalServerError))
		return
	}

	authTokenDetails := models.ClientReadableToken{
		ExpiresAt: newAuthTokenExpiration.Unix(),
		UserRoles: claims.UserRoles,
	}

	c.IndentedJSON(http.StatusOK, loginresponse{
		AuthToken:        newAuthToken,
		AuthTokenDetails: authTokenDetails,
	})
}

func (body loginrequestbody) validate() []string {
	var validationErrors []string
	const missingRequiredFieldMsg = "missing required field: %s"

	if body.Email == "" {
		validationErrors = append(validationErrors, fmt.Sprintf(missingRequiredFieldMsg, "email"))
	} else {
		_, err := mail.ParseAddress(body.Email)
		if err != nil {
			errStr := strings.ReplaceAll(err.Error(), "mail: ", "")
			validationErrors = append(validationErrors, fmt.Sprintf("invalid email: %s", errStr))
		}
	}

	if len(strings.Trim(body.Password, " ")) == 0 {
		validationErrors = append(validationErrors, fmt.Sprintf(missingRequiredFieldMsg, "password"))
	}

	return validationErrors
}

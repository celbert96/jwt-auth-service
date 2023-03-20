package routes

import (
	"fmt"
	"jwt-auth-service/controllers"
	"jwt-auth-service/models"
	"jwt-auth-service/repositories"
	"log"
	"net/http"
	"net/mail"
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
	authTokenString, err := models.MintToken(requestBody.Email, user.UserRoles, authTokenExpiration)
	if err != nil {
		log.Printf("routes > auth.go > getAuthToken > failed to mint auth token")
		c.IndentedJSON(http.StatusInternalServerError, models.ErrorResponse{ErrorMessage: "Could not mint token"})
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
	authTokenString, err := models.MintToken(addedUser.Email, addedUser.UserRoles, authTokenExpiration)
	if err != nil {
		log.Printf("routes > auth.go > register > failed to mint auth token")
		c.IndentedJSON(http.StatusInternalServerError, models.ErrorResponse{ErrorMessage: "Could not mint token"})
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

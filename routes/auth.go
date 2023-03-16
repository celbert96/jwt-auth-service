package routes

import (
	"jwt-auth-service/controllers"
	"jwt-auth-service/models"
	"jwt-auth-service/repositories"
	"log"
	"net/http"
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
	if !requestBody.validate() {
		log.Printf("routes > auth.go > login > invalid request > validation failed")
		c.IndentedJSON(http.StatusBadRequest, models.ErrResponseForHttpStatus(http.StatusBadRequest))
		return
	}

	env, ok := c.MustGet("env").(models.Env)
	if !ok {
		log.Println("routes > auth.go > login > env not accessible")
		c.IndentedJSON(http.StatusInternalServerError, models.ErrResponseForHttpStatus(http.StatusInternalServerError))
		return
	}

	repo := repositories.UserRepository{DBConn: *env.DB}
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

	repo := repositories.UserRepository{DBConn: *env.DB}
	controller := controllers.UserController{UserRepository: repo}

	_, err := controller.AddUser(user)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, models.ErrorResponse{ErrorMessage: err.Error()})
		return
	}

	authTokenExpiration := time.Now().Add(time.Minute * 30)
	authTokenString, err := models.MintToken(user.Email, user.UserRoles, authTokenExpiration)
	if err != nil {
		log.Printf("routes > auth.go > register > failed to mint auth token")
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

func (body loginrequestbody) validate() bool {
	if body.Email == "" || body.Password == "" {
		return false
	}

	return true
}

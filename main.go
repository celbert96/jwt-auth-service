package main

import (
	"database/sql"
	"jwt-auth-service/middleware"
	"jwt-auth-service/models"
	"jwt-auth-service/routes"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

func main() {
	initializeEnv()

	//initialize db
	cfg := mysql.Config{
		User:   os.Getenv("JWT_AUTH_SERVICE_DB_USER"),
		Passwd: os.Getenv("JWT_AUTH_SERVICE_DB_PASS"),
		Net:    "tcp",
		Addr:   os.Getenv("JWT_AUTH_SERVICE_DB_ADDR"),
		DBName: os.Getenv("JWT_AUTH_SERVICE_DB_NAME"),
	}

	db, err := sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		log.Fatal(err)
	}

	env := &models.Env{DB: db}

	router := gin.Default()
	router.Use(middleware.EnvMiddleware(*env))

	pubv1 := router.Group("/v1")
	routes.AddAuthRoutes(pubv1)

	router.Run(":8080")
}

func initializeEnv() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

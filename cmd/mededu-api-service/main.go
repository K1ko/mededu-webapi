package main

import (
	"log"
	"os"
	"strings"

	"github.com/K1ko/mededu-webapi/api"
	"github.com/gin-gonic/gin"
)

func main() {
	log.Printf("Server started")

	port := os.Getenv("MEDEDU_API_PORT")
	if port == "" {
		port = "8080"
	}

	environment := os.Getenv("MEDEDU_API_ENVIRONMENT")
	if !strings.EqualFold(environment, "production") {
		gin.SetMode(gin.DebugMode)
	}

	engine := gin.New()
	engine.Use(gin.Recovery())
	engine.GET("/openapi", api.HandleOpenAPI)

	if err := engine.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}

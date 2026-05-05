package main

import (
	"context"
	"log"
	"os"
	"strings"
	"time"

	"github.com/K1ko/mededu-webapi/api"
	"github.com/K1ko/mededu-webapi/internal/db_service"
	"github.com/K1ko/mededu-webapi/internal/mededu"
	"github.com/gin-contrib/cors"
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
	engine.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "PUT", "POST", "DELETE", "PATCH"},
		AllowHeaders:     []string{"Origin", "Authorization", "Content-Type"},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	}))

	dbService := db_service.NewMongoService[mededu.Training](db_service.MongoServiceConfig{})
	defer dbService.Disconnect(context.Background())
	engine.Use(func(ctx *gin.Context) {
		ctx.Set("db_service", dbService)
		ctx.Next()
	})

	handleFunctions := mededu.ApiHandleFunctions{
		DepartmentsAPI: mededu.NewDepartmentsAPI(),
		TrainingsAPI:   mededu.NewTrainingsAPI(),
	}
	mededu.NewRouterWithGinEngine(engine, handleFunctions)

	engine.GET("/openapi", api.HandleOpenAPI)

	if err := engine.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}

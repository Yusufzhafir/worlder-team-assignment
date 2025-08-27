// cmd/microservice-a/main.go
package main

import (
	"log"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/Yusufzhafir/worlder-team-assignment/a-service/router"
	"github.com/Yusufzhafir/worlder-team-assignment/a-service/router/generator"
	"github.com/Yusufzhafir/worlder-team-assignment/a-service/usecase"
)

func main() {
	// Initialize data generator
	dg := usecase.NewDataGenerator("localhost:50051")
	err := dg.Connect()
	if err != nil {
		log.Fatalf("Failed to connect to gRPC server: %v", err)
	}
	defer dg.Close()

	// Initialize Echo
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	generatorRouter := generator.NewGeneratorRouter(&dg)
	router.BindRouter(router.BindRouterOpts{
		E:      e,
		Router: generatorRouter,
	})

	log.Println("Starting Microservice A on :8080")
	e.Logger.Fatal(e.Start(":8080"))
}

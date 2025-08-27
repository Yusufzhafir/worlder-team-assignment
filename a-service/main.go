// cmd/microservice-a/main.go
package main

import (
	"errors"
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/Yusufzhafir/worlder-team-assignment/a-service/router"
	"github.com/Yusufzhafir/worlder-team-assignment/a-service/router/generator"
	"github.com/Yusufzhafir/worlder-team-assignment/a-service/usecase"
)

//	@title			WORLDER TEAM ASSIGNMENT
//	@version		1.0
//	@description	This is a sample server celler server.
//	@termsOfService	http://swagger.io/terms/

//	@contact.name	API Support
//	@contact.url	http://www.swagger.io/support
//	@contact.email	support@swagger.io

//	@license.name	Apache 2.0
//	@license.url	http://www.apache.org/licenses/LICENSE-2.0.html

//	@host		localhost:9000
//	@BasePath	/api/v1

//	@securityDefinitions.basic	BasicAuth

//	@securityDefinitions.apikey	ApiKeyAuth
//	@in							header
//	@name						Authorization
//	@description				Description for what is this security definition being used

//	@securitydefinitions.oauth2.application	OAuth2Application
//	@tokenUrl								https://example.com/oauth/token
//	@scope.write							Grants write access
//	@scope.admin							Grants read and write access to administrative information

//	@securitydefinitions.oauth2.implicit	OAuth2Implicit
//	@authorizationUrl						https://example.com/oauth/authorize
//	@scope.write							Grants write access
//	@scope.admin							Grants read and write access to administrative information

//	@securitydefinitions.oauth2.password	OAuth2Password
//	@tokenUrl								https://example.com/oauth/token
//	@scope.read								Grants read access
//	@scope.write							Grants write access
//	@scope.admin							Grants read and write access to administrative information

// @securitydefinitions.oauth2.accessCode	OAuth2AccessCode
// @tokenUrl								https://example.com/oauth/token
// @authorizationUrl						https://example.com/oauth/authorize
// @scope.admin							Grants read and write access to administrative information
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

	log.Println("Starting Microservice A on :9000")

	if err = e.Start(":9000"); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("crashed the server %v", err)
	}
	log.Println("servers stopped cleanly")
}

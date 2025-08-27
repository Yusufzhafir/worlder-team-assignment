package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	//internal
	_ "github.com/Yusufzhafir/worlder-team-assignment/b-service/docs"
	grpcServer "github.com/Yusufzhafir/worlder-team-assignment/b-service/grpc"
	sensorRepository "github.com/Yusufzhafir/worlder-team-assignment/b-service/repository"
	httpRouter "github.com/Yusufzhafir/worlder-team-assignment/b-service/router"
	sensorRouter "github.com/Yusufzhafir/worlder-team-assignment/b-service/router/sensor"
	sensorUsecase "github.com/Yusufzhafir/worlder-team-assignment/b-service/usecase"
	pb "github.com/Yusufzhafir/worlder-team-assignment/common/protobuf"
	"golang.org/x/sync/errgroup"

	//external
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	echoSwagger "github.com/swaggo/echo-swagger"
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

//	@host		localhost:8080
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

var (
	port = 50051
)

func main() {
	// ---- shared ctx (CTRL+C to stop) ----
	rootCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	logger := log.Default()
	//load environment variable
	err := godotenv.Load()
	if err != nil {
		logger.Fatal("Error loading .env file")
	}
	dbName := os.Getenv("MYSQL_DATABASE")
	dbUser := os.Getenv("MYSQL_USER")
	dbPassword := os.Getenv("MYSQL_PASSWORD")

	grpcLis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))

	if err != nil {
		logger.Fatalf("failed to listen: %v", err)
	}

	var connString string
	if val := os.Getenv("DB_DSN"); val != "" {
		connString = val
	} else {
		connString = fmt.Sprintf("%s:%s@tcp(localhost:3306)/%s?parseTime=true&loc=Local", dbUser, dbPassword, dbName)
	}
	db, err := sqlx.Connect("mysql", connString)

	if err != nil {
		logger.Fatalf("Failed to connect DB %v", err)
		return
	}

	//initiate stuff
	repoObj := sensorRepository.NewSensorRepository()
	useCaseObj := sensorUsecase.NewSensorUseCase(
		db,
		&repoObj,
	)

	//grpc server
	grpcSrv := grpc.NewServer()
	myServer := grpcServer.NewServerGRPC(grpcServer.ServerGRPCOpts{
		SensorUseCase: &useCaseObj,
		Logger:        logger,
	})
	pb.RegisterIngestServiceServer(grpcSrv, &myServer)

	//http server
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.GET("/swagger/*", echoSwagger.WrapHandler)

	//http routers
	sensorRouter := sensorRouter.NewSensorRouter(&useCaseObj)
	httpRouter.BindRouter(httpRouter.BindRouterOpts{
		E:            e,
		SensorRouter: &sensorRouter,
	})

	// ---- run both ----
	g, ctx := errgroup.WithContext(rootCtx)

	// gRPC server
	g.Go(func() error {
		log.Printf("gRPC server listening at %s", grpcLis.Addr())
		// stop gRPC when context is canceled
		go func() {
			<-ctx.Done()
			log.Println("stopping gRPC...")
			grpcSrv.GracefulStop()
		}()
		return grpcSrv.Serve(grpcLis)
	})

	// REST server
	g.Go(func() error {
		log.Printf("REST server listening at localhost:8080")
		// graceful shutdown
		go func() {
			<-ctx.Done()
			log.Println("stopping REST...")
			shCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			_ = e.Shutdown(shCtx)
		}()
		err := e.Start(":8080")
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		return nil
	})

	// wait until one server errors or context canceled
	if err := g.Wait(); err != nil && !errors.Is(err, context.Canceled) {
		log.Fatalf("server error: %v", err)
	}
	log.Println("servers stopped cleanly")
}

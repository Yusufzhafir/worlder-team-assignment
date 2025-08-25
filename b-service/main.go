package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"

	//internal
	sensorRepository "github.com/Yusufzhafir/worlder-team-assignment/b-service/repository"
	sensorUsecase "github.com/Yusufzhafir/worlder-team-assignment/b-service/usecase"
	pb "github.com/Yusufzhafir/worlder-team-assignment/common/protobuf"

	//external
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
)

var (
	port = 50051
)

type ServerGRPC struct {
	pb.UnimplementedIngestServiceServer
	sensorUsecase sensorUsecase.SensorUseCase
	logger        *log.Logger
}

// SayHello implements helloworld.GreeterServer
func (s *ServerGRPC) Readings(ctx context.Context, in *pb.SensorReading) (*pb.StreamAck, error) {
	err := s.sensorUsecase.InsertSensor(ctx, in)

	if err != nil {
		s.logger.Printf("WATDEFAK HAPPEND %v", err)
		return &pb.StreamAck{
			Status: fmt.Sprintf("failed received id %v", in),
		}, err
	}

	s.logger.Printf("Successfully saved HAPPEND %v", err)
	return &pb.StreamAck{
		Status: fmt.Sprintf("successfully received id %v", in),
	}, nil
}

func main() {
	err := godotenv.Load()
	logger := log.Default()
	if err != nil {
		logger.Fatal("Error loading .env file")
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))

	if err != nil {
		logger.Fatalf("failed to listen: %v", err)
	}

	dbName := os.Getenv("MYSQL_DATABASE")
	dbUser := os.Getenv("MYSQL_USER")
	dbPassword := os.Getenv("MYSQL_PASSWORD")
	db, err := sqlx.Connect("mysql", fmt.Sprintf("%s:%s@(localhost:3306)/%s", dbUser, dbPassword, dbName))

	if err != nil {
		logger.Fatalf("Failed to connect DB %v", err)
		return
	}

	//initiate stuff
	repoObj := &sensorRepository.SensorRepositoryImpl{}
	useCaseObj := &sensorUsecase.SensorUseCaseImpl{
		Db:   db,
		Repo: repoObj,
	}
	s := grpc.NewServer()
	myServer := &ServerGRPC{
		sensorUsecase: useCaseObj,
		logger:        logger,
	}
	pb.RegisterIngestServiceServer(s, myServer)
	logger.Printf("server listening at %v", lis.Addr())

	if err := s.Serve(lis); err != nil {
		logger.Fatalf("failed to serve: %v", err)
	}
}

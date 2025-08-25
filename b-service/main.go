package main

import (
	"context"
	"fmt"
	"log"
	"net"

	pb "github.com/Yusufzhafir/worlder-team-assignment/common/protobuf"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
)

var (
	port = 50051
)

type ServerGRPC struct {
	pb.UnimplementedIngestServiceServer
	db *sqlx.DB
}

// SayHello implements helloworld.GreeterServer
func (s *ServerGRPC) Readings(_ context.Context, in *pb.SensorReading) (*pb.StreamAck, error) {
	log.Printf("Received: %v", in.GetId1())
	return &pb.StreamAck{
		Status: fmt.Sprintf("successfully received id %v", in),
	}, nil
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	db, err := sqlx.Connect("postgres", "user=foo dbname=bar sslmode=disable")

	if err != nil {
		log.Fatalf("Failed to connect DB %v", err)
		return
	}

	s := grpc.NewServer()
	myServer := &ServerGRPC{
		db: db,
	}
	pb.RegisterIngestServiceServer(s, myServer)
	log.Printf("server listening at %v", lis.Addr())

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

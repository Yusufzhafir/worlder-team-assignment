package main

import (
	"context"
	"fmt"
	"log"
	"net"

	pb "github.com/Yusufzhafir/worlder-team-assignment/common/protobuf"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
)

var (
	port = 50051
)

type ServerGRPC struct {
	pb.UnimplementedIngestServiceServer
}

// SayHello implements helloworld.GreeterServer
func (s *ServerGRPC) StreamReadings(_ context.Context, in *pb.SensorReading) (*pb.StreamAck, error) {
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
	s := grpc.NewServer()
	pb.RegisterIngestServiceServer(s, &ServerGRPC{})
	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

package main

import (
	"context"
	"log"
	"time"

	pb "github.com/Yusufzhafir/worlder-team-assignment/common/protobuf"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	address = "localhost:50051"
)

func main() {
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
		return
	}
	defer conn.Close()
	c := pb.NewIngestServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	ack, err := c.Readings(ctx, &pb.SensorReading{
		Value:       10,
		SensorType:  "manuel",
		Id1:         "asldk",
		Id2:         10,
		TimestampMs: time.Now().Unix(),
	})

	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	log.Printf("Status Result: %s", ack.GetStatus())
}

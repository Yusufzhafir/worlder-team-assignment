package grpc

import (
	"context"
	"fmt"
	"log"

	sensorUsecase "github.com/Yusufzhafir/worlder-team-assignment/b-service/usecase"
	pb "github.com/Yusufzhafir/worlder-team-assignment/common/protobuf"
)

type ServerGRPC struct {
	pb.UnimplementedIngestServiceServer
	sensorUsecase *sensorUsecase.SensorUseCase
	logger        *log.Logger
}

// SayHello implements helloworld.GreeterServer
func (s *ServerGRPC) Readings(ctx context.Context, in *pb.SensorReading) (*pb.StreamAck, error) {
	usecase := *s.sensorUsecase

	if usecase == nil {
		return &pb.StreamAck{
			Status: fmt.Sprintf("usecase is nil %v", usecase),
		}, fmt.Errorf("usecase is nil %v", usecase)
	}

	err := usecase.InsertSensor(ctx, in)

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

type ServerGRPCOpts struct {
	SensorUseCase *sensorUsecase.SensorUseCase
	Logger        *log.Logger
}

func NewServerGRPC(opts ServerGRPCOpts) ServerGRPC {
	return ServerGRPC{
		sensorUsecase: opts.SensorUseCase,
		logger:        opts.Logger,
	}
}

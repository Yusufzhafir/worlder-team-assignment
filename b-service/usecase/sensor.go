package usecase

import (
	"context"
	"fmt"
	"time"

	repository "github.com/Yusufzhafir/worlder-team-assignment/b-service/repository"
	"github.com/Yusufzhafir/worlder-team-assignment/b-service/repository/model"
	pb "github.com/Yusufzhafir/worlder-team-assignment/common/protobuf"
	"github.com/jmoiron/sqlx"
)

type SensorUseCase interface {
	InsertSensor(ctx context.Context, data *pb.SensorReading) error
}

type SensorUseCaseImpl struct {
	Db   *sqlx.DB
	Repo repository.SensorRepository
}

func (sensorUseCase *SensorUseCaseImpl) InsertSensor(ctx context.Context, data *pb.SensorReading) error {
	id, err := sensorUseCase.Repo.InsertReadingTx(ctx, sensorUseCase.Db, &model.SensorReadingInsert{
		SensorValue: data.GetValue(),
		SensorType:  data.GetSensorType(),
		ID1:         data.GetId1(),
		ID2:         int(data.GetId2()),
		TS:          time.Unix(data.GetTimestampMs(), 0),
	})

	if err != nil {
		return err
	}
	if id == 0 {
		return fmt.Errorf("failed to insert because id returned with %d", id)
	}
	return nil
}

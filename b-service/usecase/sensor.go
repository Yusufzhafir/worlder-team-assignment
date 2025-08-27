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
	GetSensorByTime(ctx context.Context, from time.Time, to time.Time, limit int, offset int) (getSensorByType, error)
}

type SensorUseCaseImpl struct {
	db   *sqlx.DB
	repo *repository.SensorRepository
}

func (sensorUseCase *SensorUseCaseImpl) InsertSensor(ctx context.Context, data *pb.SensorReading) error {
	repo := *sensorUseCase.repo

	if repo == nil {
		return fmt.Errorf("repository object is nil %v", repo)
	}

	id, err := repo.InsertReadingTx(ctx, sensorUseCase.db, &model.SensorReadingInsert{
		SensorValue: data.GetValue(),
		SensorType:  data.GetSensorType(),
		ID1:         data.GetId1(),
		ID2:         int(data.GetId2()),
		TS:          time.UnixMilli(data.GetTimestampMs()),
	})

	if err != nil {
		return err
	}

	if id == 0 {
		return fmt.Errorf("failed to insert because id returned with %d", id)
	}

	return nil
}

type getSensorByType struct {
	Data  []model.SensorReading
	Count int64
}

func (sensorUseCase *SensorUseCaseImpl) GetSensorByTime(ctx context.Context, from time.Time, to time.Time, limit int, offset int) (getSensorByType, error) {
	repo := *sensorUseCase.repo
	result := getSensorByType{}

	if repo == nil {
		return result, fmt.Errorf("repository object is nil %v", repo)
	}

	rows, err := repo.SelectByTime(ctx, sensorUseCase.db, from, to, limit, offset)
	if err != nil {
		return result, err
	}
	count, err := repo.SelectCountByTime(ctx, sensorUseCase.db, from, to)
	if err != nil {
		return result, err
	}

	result.Count = count
	result.Data = rows

	return result, nil
}

func NewSensorUseCase(
	db *sqlx.DB,
	repo *repository.SensorRepository,
) SensorUseCase {
	return &SensorUseCaseImpl{
		db:   db,
		repo: repo,
	}
}

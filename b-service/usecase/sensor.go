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
	GetSensorPaginated(ctx context.Context, limit int, offset int) (paginatedSensor, error)
	GetSensorByTime(ctx context.Context, from time.Time, to time.Time, limit int, offset int) (paginatedSensor, error)
	GetSensorByIDs(ctx context.Context, idCombinationPtr *[]repository.IDCombination, limit int, offset int) (paginatedSensor, error)
	GetSensorByIDsAndTime(ctx context.Context, idCombinationPtr *[]repository.IDCombination, from time.Time, to time.Time, limit int, offset int) (paginatedSensor, error)
	DeleteSensorByIds(ctx context.Context, idCombinationPtr *[]repository.IDCombination) (mutatedResponse, error)
	DeleteSensorByTime(ctx context.Context, from time.Time, to time.Time) (mutatedResponse, error)
	DeleteSensorByIdsAndTime(ctx context.Context, idCombinationPtr *[]repository.IDCombination, from time.Time, to time.Time) (mutatedResponse, error)
	UpdateSensorByIds(ctx context.Context, idCombinationPtr *[]repository.IDCombination, sensorValue float64, sensorType string) (mutatedResponse, error)
	UpdateSensorByTime(ctx context.Context, from time.Time, to time.Time, sensorValue float64, sensorType string) (mutatedResponse, error)
	UpdateSensorByIdsAndTime(ctx context.Context, idCombinationPtr *[]repository.IDCombination, from time.Time, to time.Time, sensorValue float64, sensorType string) (mutatedResponse, error)
}

type SensorUseCaseImpl struct {
	db   *sqlx.DB
	repo *repository.SensorRepository
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

type paginatedSensor struct {
	Data  []model.SensorReading
	Count int64
}

func (sensorUseCase *SensorUseCaseImpl) GetSensorByTime(ctx context.Context, from time.Time, to time.Time, limit int, offset int) (paginatedSensor, error) {
	repo := *sensorUseCase.repo
	result := paginatedSensor{}

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

func (sensorUseCase *SensorUseCaseImpl) GetSensorPaginated(ctx context.Context, limit int, offset int) (paginatedSensor, error) {
	repo := *sensorUseCase.repo
	result := paginatedSensor{}
	if repo == nil {
		return result, fmt.Errorf("repository object is nil %v", repo)
	}

	rows, err := repo.SelectSensorDataPaginated(ctx, sensorUseCase.db, limit, offset)
	if err != nil {
		return result, err
	}
	count, err := repo.SelectCountPagination(ctx, sensorUseCase.db)
	if err != nil {
		return result, err
	}

	result.Count = count
	result.Data = rows

	return result, nil
}

func (sensorUseCase *SensorUseCaseImpl) GetSensorByIDs(ctx context.Context, idCombinationPtr *[]repository.IDCombination, limit int, offset int) (paginatedSensor, error) {
	repo := *sensorUseCase.repo
	result := paginatedSensor{}
	idCombination := *idCombinationPtr
	if repo == nil {
		return result, fmt.Errorf("repository object is nil %v", repo)
	}

	rows, err := repo.SelectByIDs(ctx, sensorUseCase.db, idCombination, limit, offset)
	if err != nil {
		return result, err
	}
	count, err := repo.SelectCountByIDs(ctx, sensorUseCase.db, idCombination)
	if err != nil {
		return result, err
	}

	result.Count = count
	result.Data = rows

	return result, nil
}

func (sensorUseCase *SensorUseCaseImpl) GetSensorByIDsAndTime(ctx context.Context, idCombinationPtr *[]repository.IDCombination, from time.Time, to time.Time, limit int, offset int) (paginatedSensor, error) {
	repo := *sensorUseCase.repo
	result := paginatedSensor{}
	idCombination := *idCombinationPtr
	if repo == nil {
		return result, fmt.Errorf("repository object is nil %v", repo)
	}

	rows, err := repo.SelectByIDsAndTime(ctx, sensorUseCase.db, idCombination, from, to, limit, offset)
	if err != nil {
		return result, err
	}
	count, err := repo.SelectCountByIDsAndTime(ctx, sensorUseCase.db, idCombination, from, to)
	if err != nil {
		return result, err
	}

	result.Count = count
	result.Data = rows

	return result, nil
}

type mutatedResponse struct {
	Message string
	Count   int64
}

func (sensorUseCase *SensorUseCaseImpl) DeleteSensorByIds(ctx context.Context, idCombinationPtr *[]repository.IDCombination) (mutatedResponse, error) {
	repo := *sensorUseCase.repo
	result := mutatedResponse{}
	idCombination := *idCombinationPtr
	if repo == nil {
		return result, fmt.Errorf("repository object is nil %v", repo)
	}

	rows, err := repo.DeleteByIDs(ctx, sensorUseCase.db, idCombination)
	if err != nil {
		return result, err
	}

	result.Count = rows
	result.Message = "Successfully deleted rows by Ids"

	return result, nil
}

func (sensorUseCase *SensorUseCaseImpl) DeleteSensorByTime(ctx context.Context, from time.Time, to time.Time) (mutatedResponse, error) {
	repo := *sensorUseCase.repo
	result := mutatedResponse{}

	if repo == nil {
		return result, fmt.Errorf("repository object is nil %v", repo)
	}

	rows, err := repo.DeleteByTime(ctx, sensorUseCase.db, from, to)
	if err != nil {
		return result, err
	}

	result.Count = rows
	result.Message = "Successfully deleted rows by Time"

	return result, nil
}

func (sensorUseCase *SensorUseCaseImpl) DeleteSensorByIdsAndTime(ctx context.Context, idCombinationPtr *[]repository.IDCombination, from time.Time, to time.Time) (mutatedResponse, error) {
	repo := *sensorUseCase.repo
	result := mutatedResponse{}
	idCombination := *idCombinationPtr
	if repo == nil {
		return result, fmt.Errorf("repository object is nil %v", repo)
	}

	rows, err := repo.DeleteByIDsAndTime(ctx, sensorUseCase.db, idCombination, from, to)
	if err != nil {
		return result, err
	}

	result.Count = rows
	result.Message = "Successfully deleted rows by Ids and Time"

	return result, nil
}

func (sensorUseCase *SensorUseCaseImpl) UpdateSensorByIds(ctx context.Context, idCombinationPtr *[]repository.IDCombination, sensorValue float64, sensorType string) (mutatedResponse, error) {
	repo := *sensorUseCase.repo
	result := mutatedResponse{}
	idCombination := *idCombinationPtr
	if repo == nil {
		return result, fmt.Errorf("repository object is nil %v", repo)
	}

	rows, err := repo.UpdateByIDs(ctx, sensorUseCase.db, idCombination, sensorValue, sensorType)
	if err != nil {
		return result, err
	}

	result.Count = rows
	result.Message = "Successfully updated rows by Ids"

	return result, nil
}

func (sensorUseCase *SensorUseCaseImpl) UpdateSensorByTime(ctx context.Context, from time.Time, to time.Time, sensorValue float64, sensorType string) (mutatedResponse, error) {
	repo := *sensorUseCase.repo
	result := mutatedResponse{}

	if repo == nil {
		return result, fmt.Errorf("repository object is nil %v", repo)
	}

	rows, err := repo.UpdateByTime(ctx, sensorUseCase.db, from, to, sensorValue, sensorType)
	if err != nil {
		return result, err
	}

	result.Count = rows
	result.Message = "Successfully updated rows by Time"

	return result, nil
}

func (sensorUseCase *SensorUseCaseImpl) UpdateSensorByIdsAndTime(ctx context.Context, idCombinationPtr *[]repository.IDCombination, from time.Time, to time.Time, sensorValue float64, sensorType string) (mutatedResponse, error) {
	repo := *sensorUseCase.repo
	result := mutatedResponse{}
	idCombination := *idCombinationPtr
	if repo == nil {
		return result, fmt.Errorf("repository object is nil %v", repo)
	}

	rows, err := repo.UpdateByIDsAndTime(ctx, sensorUseCase.db, idCombination, from, to, sensorValue, sensorType)
	if err != nil {
		return result, err
	}

	result.Count = rows
	result.Message = "Successfully updated rows by Ids and Time"

	return result, nil
}

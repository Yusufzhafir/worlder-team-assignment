package router

import (
	"github.com/Yusufzhafir/worlder-team-assignment/b-service/usecase"
	"github.com/labstack/echo/v4"
)

type SensorRouter interface {
	GetSensorDataById(ctx echo.Context) error
	GetSensorDataByTime(ctx echo.Context) error
	GetSensorDataByIdAndTime(ctx echo.Context) error
	DeleteSensorById(ctx echo.Context) error
	DeleteSensorByTime(ctx echo.Context) error
	DeleteSensorByIdAndTime(ctx echo.Context) error
	UpdateSensorById(ctx echo.Context) error
	UpdateSensorByTime(ctx echo.Context) error
	UpdateSensorByIdAndTime(ctx echo.Context) error
	GetSensorDataPaginated(ctx echo.Context) error
}

type SensorRouterImpl struct {
	sensorUsecase *usecase.SensorUseCase
}

func NewSensorRouter(sensorUsecase *usecase.SensorUseCase) SensorRouter {
	return &SensorRouterImpl{
		sensorUsecase: sensorUsecase,
	}
}
func (s *SensorRouterImpl) GetSensorDataById(ctx echo.Context) error {
	return nil
}

func (s *SensorRouterImpl) GetSensorDataByTime(ctx echo.Context) error {
	return nil
}

func (s *SensorRouterImpl) GetSensorDataByIdAndTime(ctx echo.Context) error {
	return nil
}

func (s *SensorRouterImpl) DeleteSensorById(ctx echo.Context) error {
	return nil
}

func (s *SensorRouterImpl) DeleteSensorByTime(ctx echo.Context) error {
	return nil
}

func (s *SensorRouterImpl) DeleteSensorByIdAndTime(ctx echo.Context) error {
	return nil
}

func (s *SensorRouterImpl) UpdateSensorById(ctx echo.Context) error {
	return nil
}

func (s *SensorRouterImpl) UpdateSensorByTime(ctx echo.Context) error {
	return nil
}

func (s *SensorRouterImpl) UpdateSensorByIdAndTime(ctx echo.Context) error {
	return nil
}

func (s *SensorRouterImpl) GetSensorDataPaginated(ctx echo.Context) error {
	return nil
}

package router

import (
	"net/http"
	"time"

	payload "github.com/Yusufzhafir/worlder-team-assignment/b-service/router/sensor/model"
	httpmodels "github.com/Yusufzhafir/worlder-team-assignment/b-service/shared/model"
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

// GetSensorDataPaginated godoc
// @Summary     List sensor readings (paginated)
// @Tags        sensor
// @Accept      json
// @Produce     json
// @Param       page       query   int    false  "Page number"  minimum(1) default(1)
// @Param       page_size  query   int    false  "Page size"    minimum(1) maximum(500) default(50)
// @Success     200 {object} model.Envelope{data=model.SensorPage} "data: SensorPage"
// @Failure     400 {object} model.Envelope{data=model.Empty} "error=true, message explains"
// @Failure     500 {object} model.Envelope{data=model.Empty}
// @Router      /sensor [get]
func (s *SensorRouterImpl) GetSensorDataPaginated(ctx echo.Context) error {
	page := ctx.QueryParams().Get("page")
	pageSize := ctx.QueryParams().Get("page_size")
	if page == "" || pageSize == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Please provide valid query param")
	}
	body := httpmodels.Body[payload.SensorPayload]{
		Data: payload.SensorPayload{
			ID:          "10",
			SensorType:  "alksjd",
			Value:       10,
			TimestampMs: time.Now().Unix(),
		},
	}
	return ctx.JSON(http.StatusOK, body)
}

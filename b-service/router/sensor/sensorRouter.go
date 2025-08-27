package router

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Yusufzhafir/worlder-team-assignment/b-service/repository"
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

func validatePagination(ctx echo.Context) (int, int, int, error) {
	pageStr := ctx.QueryParam("page")
	sizeStr := ctx.QueryParam("page_size")
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		return 0, 0, 0, ctx.JSON(http.StatusBadRequest, httpmodels.Body[httpmodels.Empty]{Error: true, Message: "page must be a positive integer"})
	}
	size, err := strconv.Atoi(sizeStr)
	if err != nil || size < 1 || size > 500 {
		return 0, 0, 0, ctx.JSON(http.StatusBadRequest, httpmodels.Body[httpmodels.Empty]{Error: true, Message: "page_size must be 1..500"})
	}
	offset := (page - 1) * size
	return page, size, offset, nil
}

// GetSensorDataByIds godoc
// @Summary     List sensor readings filtered by ID combinations (paginated)
// @Tags        sensor
// @Accept      json
// @Produce     json
// @Param       page       query   int    false  "Page number"              minimum(1) default(1)
// @Param       page_size  query   int    false  "Page size"                minimum(1) maximum(500) default(50)
// @Param       id1        query   string false  "Comma-separated ID1 values" example(0,1,2)
// @Param       id2        query   string false  "Comma-separated ID2 values" example(A,B,C)
// @Success     200 {object} model.Envelope{data=model.SensorPage} "data: SensorPage"
// @Failure     400 {object} model.Envelope{data=model.Empty} "error=true, message explains"
// @Failure     500 {object} model.Envelope{data=model.Empty}
// @Router      /sensor/ids [get]
func (s *SensorRouterImpl) GetSensorDataById(ctx echo.Context) error {
	usecase := *s.sensorUsecase
	if usecase == nil {
		return ctx.JSON(http.StatusInternalServerError, httpmodels.Body[httpmodels.Empty]{
			Error:   true,
			Message: "usecase is not provided",
		})
	}

	// pagination
	page, size, offset, err := validatePagination(ctx)
	if err != nil {
		return err
	}

	// Parse ID combinations from query parameters
	idCombinations, err := parseIDCombinations(ctx)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, httpmodels.Body[httpmodels.Empty]{
			Error:   true,
			Message: fmt.Sprintf("Invalid ID parameters: %v", err),
		})
	}

	if len(idCombinations) == 0 {
		return ctx.JSON(http.StatusBadRequest, httpmodels.Body[httpmodels.Empty]{
			Error:   true,
			Message: "At least one ID combination must be provided",
		})
	}

	// call usecase
	result, err := usecase.GetSensorByID(ctx.Request().Context(), &idCombinations, size, offset)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, httpmodels.Body[httpmodels.Empty]{
			Error:   true,
			Message: err.Error(),
		})
	}

	newPayload := make([]payload.SensorPayload, len(result.Data))
	for i := 0; i < len(result.Data); i++ {
		currElement := result.Data[i]
		newPayload[i] = payload.SensorPayload{
			ID1:         currElement.ID1,
			ID2:         currElement.ID2,
			SensorType:  currElement.SensorType,
			Value:       currElement.SensorValue,
			TimestampMs: currElement.TS.Unix(),
		}
	}

	body := httpmodels.Body[payload.SensorPage]{
		Data: payload.SensorPage{
			Items:    newPayload,
			Page:     page,
			PageSize: size,
			Total:    result.Count,
		},
	}
	return ctx.JSON(http.StatusOK, body)
}

// parseIDCombinations parses ID combinations from query parameters
// Supports both comma-separated arrays and single values
// Examples:
//   - ?id1=0&id2=A (single combination)
//   - ?id1=0,1,2&id2=A,B,C (multiple combinations)
func parseIDCombinations(ctx echo.Context) ([]repository.IDCombination, error) {
	id1Str := ctx.QueryParam("id1")
	id2Str := ctx.QueryParam("id2")

	if id1Str == "" && id2Str == "" {
		return []repository.IDCombination{}, nil
	}

	if id1Str == "" || id2Str == "" {
		return nil, fmt.Errorf("both id1 and id2 must be provided")
	}

	// Split comma-separated values
	id1Parts := strings.Split(id1Str, ",")
	id2Parts := strings.Split(id2Str, ",")

	if len(id1Parts) != len(id2Parts) {
		return nil, fmt.Errorf("number of id1 values (%d) must match number of id2 values (%d)", len(id1Parts), len(id2Parts))
	}

	var combinations []repository.IDCombination
	for i := 0; i < len(id1Parts); i++ {
		// Trim whitespace
		id1Part := strings.TrimSpace(id1Parts[i])
		id2Part := strings.TrimSpace(id2Parts[i])

		// Convert ID1 to int
		id2Int, err := strconv.Atoi(id2Part)
		if err != nil {
			return nil, fmt.Errorf("invalid id1 value '%s': must be integer", id1Part)
		}

		combinations = append(combinations, repository.IDCombination{
			ID1: id1Part,
			ID2: id2Int,
		})
	}

	return combinations, nil
}

// GetSensorDataByTime godoc
// @Summary     List sensor readings filter by time(paginated)
// @Tags        sensor
// @Accept      json
// @Produce     json
// @Param       page       query   int    false  "Page number"  minimum(1) default(1)
// @Param       page_size  query   int    false  "Page size"    minimum(1) maximum(500) default(50)
// @Param       from       query   string    false  "from time"  default(2006-01-02T15:04:05.999999999+07:00)
// @Param       to  	query   string    false  "to time"    default(2006-01-02T15:04:05.999999999+07:00)
// @Success     200 {object} model.Envelope{data=model.SensorPage} "data: SensorPage"
// @Failure     400 {object} model.Envelope{data=model.Empty} "error=true, message explains"
// @Failure     500 {object} model.Envelope{data=model.Empty}
// @Router      /sensor/time [get]
func (s *SensorRouterImpl) GetSensorDataByTime(ctx echo.Context) error {
	usecase := *s.sensorUsecase
	if usecase == nil {
		return ctx.JSON(http.StatusInternalServerError, httpmodels.Body[httpmodels.Empty]{
			Error:   true,
			Message: "usecase is not provided",
		})
	}

	// pagination
	page, size, offset, err := validatePagination(ctx)
	if err != nil {
		return err
	}

	// time range
	fromStr := ctx.QueryParam("from")
	toStr := ctx.QueryParam("to")
	from, err := time.Parse(time.RFC3339Nano, fromStr)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, httpmodels.Body[httpmodels.Empty]{Error: true, Message: err.Error()})
	}
	to, err := time.Parse(time.RFC3339Nano, toStr)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, httpmodels.Body[httpmodels.Empty]{Error: true, Message: err.Error()})
	}
	if !from.Before(to) {
		return ctx.JSON(http.StatusBadRequest, httpmodels.Body[httpmodels.Empty]{Error: true, Message: "`from` must be earlier than `to`"})
	}

	// call usecase
	result, err := usecase.GetSensorByTime(ctx.Request().Context(), from, to, size, offset)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, httpmodels.Body[httpmodels.Empty]{Error: true, Message: err.Error()})
	}
	newPayload := make([]payload.SensorPayload, len(result.Data))
	for i := 0; i < len(result.Data); i++ {
		currElement := result.Data[i]
		newPayload[i] = payload.SensorPayload{
			ID1:         currElement.ID1,
			ID2:         currElement.ID2,
			SensorType:  currElement.SensorType,
			Value:       currElement.SensorValue,
			TimestampMs: currElement.TS.Unix(),
		}
	}
	body := httpmodels.Body[payload.SensorPage]{
		Data: payload.SensorPage{
			Items:    newPayload,
			Page:     page,
			PageSize: size,
			Total:    result.Count,
		},
	}
	return ctx.JSON(http.StatusOK, body)
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
			ID1:         "10",
			SensorType:  "alksjd",
			Value:       10,
			TimestampMs: time.Now().Unix(),
		},
	}
	return ctx.JSON(http.StatusOK, body)
}

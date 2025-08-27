package router

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Yusufzhafir/worlder-team-assignment/b-service/repository"
	"github.com/Yusufzhafir/worlder-team-assignment/b-service/router/sensor/model"
	httpmodels "github.com/Yusufzhafir/worlder-team-assignment/b-service/shared/model"
	"github.com/Yusufzhafir/worlder-team-assignment/b-service/usecase"
	"github.com/labstack/echo/v4"
)

type SensorRouter interface {
	GetSensorDataByIds(ctx echo.Context) error
	GetSensorDataByTime(ctx echo.Context) error
	GetSensorDataByIdsAndTime(ctx echo.Context) error
	DeleteSensorByIds(ctx echo.Context) error
	DeleteSensorByTime(ctx echo.Context) error
	DeleteSensorByIdsAndTime(ctx echo.Context) error
	UpdateSensorByIds(ctx echo.Context) error
	UpdateSensorByTime(ctx echo.Context) error
	UpdateSensorByIdsAndTime(ctx echo.Context) error
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

// parseTimeRange parses time range from query parameters
func parseTimeRange(ctx echo.Context) (time.Time, time.Time, error) {
	fromStr := ctx.QueryParam("from")
	toStr := ctx.QueryParam("to")

	if fromStr == "" || toStr == "" {
		return time.Time{}, time.Time{}, fmt.Errorf("both from_time and to_time must be provided")
	}

	fromTime, err := time.Parse(time.RFC3339Nano, fromStr)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid from_time format: %v (use RFC3339 format like 2023-01-01T00:00:00Z)", err)
	}

	toTime, err := time.Parse(time.RFC3339Nano, toStr)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid to_time format: %v (use RFC3339 format like 2023-01-01T00:00:00Z)", err)
	}

	if !fromTime.Before(toTime) {
		return time.Time{}, time.Time{}, fmt.Errorf("from_time must be before to_time")
	}

	log.Default().Printf("this is converted millis converted %f, %f", fromTime.UnixMilli(), toTime.UnixMilli())

	return fromTime, toTime, nil
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

// GetSensorDataByIdss godoc
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
func (s *SensorRouterImpl) GetSensorDataByIds(ctx echo.Context) error {
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
	result, err := usecase.GetSensorByIDs(ctx.Request().Context(), &idCombinations, size, offset)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, httpmodels.Body[httpmodels.Empty]{
			Error:   true,
			Message: err.Error(),
		})
	}

	newPayload := make([]model.SensorPayload, len(result.Data))
	for i := 0; i < len(result.Data); i++ {
		currElement := result.Data[i]
		newPayload[i] = model.SensorPayload{
			ID1:         currElement.ID1,
			ID2:         currElement.ID2,
			SensorType:  currElement.SensorType,
			Value:       currElement.SensorValue,
			TimestampMs: currElement.TS.Unix(),
		}
	}

	body := httpmodels.Body[model.SensorPage]{
		Data: model.SensorPage{
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
// @Param       to  	query   string    false  "to time"    default(2006-01-02T16:04:05.999999999+07:00)
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
	fromTime, toTime, err := parseTimeRange(ctx)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, httpmodels.Body[httpmodels.Empty]{
			Error:   true,
			Message: fmt.Sprintf("Invalid time parameters: %v", err),
		})
	}

	// call usecase
	result, err := usecase.GetSensorByTime(ctx.Request().Context(), fromTime, toTime, size, offset)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, httpmodels.Body[httpmodels.Empty]{Error: true, Message: err.Error()})
	}
	newPayload := make([]model.SensorPayload, len(result.Data))
	for i := 0; i < len(result.Data); i++ {
		currElement := result.Data[i]
		newPayload[i] = model.SensorPayload{
			ID1:         currElement.ID1,
			ID2:         currElement.ID2,
			SensorType:  currElement.SensorType,
			Value:       currElement.SensorValue,
			TimestampMs: currElement.TS.Unix(),
		}
	}
	body := httpmodels.Body[model.SensorPage]{
		Data: model.SensorPage{
			Items:    newPayload,
			Page:     page,
			PageSize: size,
			Total:    result.Count,
		},
	}
	return ctx.JSON(http.StatusOK, body)
}

// GetSensorDataByIdsAndTime godoc
// @Summary     List sensor readings filtered by ID combinations and time range (paginated)
// @Tags        sensor
// @Accept      json
// @Produce     json
// @Param       page       query   int    false  "Page number"              minimum(1) default(1)
// @Param       page_size  query   int    false  "Page size"                minimum(1) maximum(500) default(50)
// @Param       id1        query   string false  "Comma-separated ID1 values" example(0,1,2)
// @Param       id2        query   string false  "Comma-separated ID2 values" example(A,B,C)
// @Param       from  query   string true   "Start time (RFC3339Nano)"     example(2025-08-25T18:00:24.947000+07:00)
// @Param       to    query   string true   "End time (RFC3339Nano)"       example(2025-08-25T19:00:24.947000+07:00)
// @Success     200 {object} model.Envelope{data=model.SensorPage} "data: SensorPage"
// @Failure     400 {object} model.Envelope{data=model.Empty} "error=true, message explains"
// @Failure     500 {object} model.Envelope{data=model.Empty}
// @Router      /sensor/ids-time [get]
func (s *SensorRouterImpl) GetSensorDataByIdsAndTime(ctx echo.Context) error {
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

	// Parse ID combinations
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

	// Parse time range
	fromTime, toTime, err := parseTimeRange(ctx)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, httpmodels.Body[httpmodels.Empty]{
			Error:   true,
			Message: fmt.Sprintf("Invalid time parameters: %v", err),
		})
	}

	// call usecase
	result, err := usecase.GetSensorByIDsAndTime(ctx.Request().Context(), &idCombinations, fromTime, toTime, size, offset)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, httpmodels.Body[httpmodels.Empty]{
			Error:   true,
			Message: err.Error(),
		})
	}

	newPayload := make([]model.SensorPayload, len(result.Data))
	for i := 0; i < len(result.Data); i++ {
		currElement := result.Data[i]
		newPayload[i] = model.SensorPayload{
			ID1:         currElement.ID1,
			ID2:         currElement.ID2,
			SensorType:  currElement.SensorType,
			Value:       currElement.SensorValue,
			TimestampMs: currElement.TS.Unix(),
		}
	}

	body := httpmodels.Body[model.SensorPage]{
		Data: model.SensorPage{
			Items:    newPayload,
			Page:     page,
			PageSize: size,
			Total:    result.Count,
		},
	}
	return ctx.JSON(http.StatusOK, body)
}

// DeleteSensorByIds godoc
// @Summary     Delete sensor readings by ID combinations
// @Description Delete sensor readings that match the specified ID combinations
// @Tags        sensor
// @Accept      json
// @Produce     json
// @Param       request body model.DeleteByIDsRequest true "ID combinations to delete"
// @Success     200 {object} model.Envelope{data=model.DeleteResponse} "data: DeleteResponse"
// @Failure     400 {object} model.Envelope{data=model.Empty} "error=true, message explains"
// @Failure     500 {object} model.Envelope{data=model.Empty}
// @Router      /sensor/delete/ids [delete]
func (s *SensorRouterImpl) DeleteSensorByIds(ctx echo.Context) error {
	usecase := *s.sensorUsecase
	if usecase == nil {
		return ctx.JSON(http.StatusInternalServerError, httpmodels.Body[httpmodels.Empty]{
			Error:   true,
			Message: "usecase is not provided",
		})
	}

	// Parse JSON request body
	var req model.DeleteByIDsRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, httpmodels.Body[httpmodels.Empty]{
			Error:   true,
			Message: fmt.Sprintf("Invalid request body: %v", err),
		})
	}

	// Validate request
	if len(req.IDCombinations) == 0 {
		return ctx.JSON(http.StatusBadRequest, httpmodels.Body[httpmodels.Empty]{
			Error:   true,
			Message: "At least one ID combination must be provided",
		})
	}

	// Convert to repository model
	var idCombinations []repository.IDCombination
	for _, combo := range req.IDCombinations {
		idCombinations = append(idCombinations, repository.IDCombination{
			ID1: combo.ID1,
			ID2: combo.ID2,
		})
	}

	// Call usecase
	deletedCount, err := usecase.DeleteSensorByIds(ctx.Request().Context(), &idCombinations)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, httpmodels.Body[httpmodels.Empty]{
			Error:   true,
			Message: err.Error(),
		})
	}

	body := httpmodels.Body[model.DeleteResponse]{
		Data: model.DeleteResponse{
			DeletedCount: deletedCount.Count,
			Message:      deletedCount.Message,
		},
	}
	return ctx.JSON(http.StatusOK, body)
}

// DeleteSensorByTime godoc
// @Summary     Delete sensor readings by time range
// @Description Delete sensor readings that fall within the specified time range
// @Tags        sensor
// @Accept      json
// @Produce     json
// @Param       request body model.DeleteByTimeRequest true "Time range to delete"
// @Success     200 {object} model.Envelope{data=model.DeleteResponse} "data: DeleteResponse"
// @Failure     400 {object} model.Envelope{data=model.Empty} "error=true, message explains"
// @Failure     500 {object} model.Envelope{data=model.Empty}
// @Router      /sensor/delete/time [delete]
func (s *SensorRouterImpl) DeleteSensorByTime(ctx echo.Context) error {
	usecase := *s.sensorUsecase
	if usecase == nil {
		return ctx.JSON(http.StatusInternalServerError, httpmodels.Body[httpmodels.Empty]{
			Error:   true,
			Message: "usecase is not provided",
		})
	}

	// Parse JSON request body
	var req model.DeleteByTimeRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, httpmodels.Body[httpmodels.Empty]{
			Error:   true,
			Message: fmt.Sprintf("Invalid request body: %v", err),
		})
	}

	// Validate time range
	if req.FromTime.IsZero() || req.ToTime.IsZero() {
		return ctx.JSON(http.StatusBadRequest, httpmodels.Body[httpmodels.Empty]{
			Error:   true,
			Message: "Both from_time and to_time must be provided",
		})
	}

	if !req.FromTime.Before(req.ToTime) {
		return ctx.JSON(http.StatusBadRequest, httpmodels.Body[httpmodels.Empty]{
			Error:   true,
			Message: "from_time must be before to_time",
		})
	}

	// Call usecase
	deletedCount, err := usecase.DeleteSensorByTime(ctx.Request().Context(), req.FromTime, req.ToTime)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, httpmodels.Body[httpmodels.Empty]{
			Error:   true,
			Message: err.Error(),
		})
	}

	body := httpmodels.Body[model.DeleteResponse]{
		Data: model.DeleteResponse{
			DeletedCount: deletedCount.Count,
			Message:      deletedCount.Message,
		},
	}
	return ctx.JSON(http.StatusOK, body)
}

// DeleteSensorByTime godoc
// @Summary     Delete sensor readings by time range
// @Description Delete sensor readings that fall within the specified time range
// @Tags        sensor
// @Accept      json
// @Produce     json
// @Param       request body model.DeleteByIDAndTimesRequest true "Time range to delete"
// @Success     200 {object} model.Envelope{data=model.DeleteResponse} "data: DeleteResponse"
// @Failure     400 {object} model.Envelope{data=model.Empty} "error=true, message explains"
// @Failure     500 {object} model.Envelope{data=model.Empty}
// @Router      /sensor/delete/ids-time [delete]
func (s *SensorRouterImpl) DeleteSensorByIdsAndTime(ctx echo.Context) error {
	usecase := *s.sensorUsecase
	if usecase == nil {
		return ctx.JSON(http.StatusInternalServerError, httpmodels.Body[httpmodels.Empty]{
			Error:   true,
			Message: "usecase is not provided",
		})
	}

	// Parse JSON request body
	var req model.DeleteByIDAndTimesRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, httpmodels.Body[httpmodels.Empty]{
			Error:   true,
			Message: fmt.Sprintf("Invalid request body: %v", err),
		})
	}

	// Validate time range
	if req.FromTime.IsZero() || req.ToTime.IsZero() {
		return ctx.JSON(http.StatusBadRequest, httpmodels.Body[httpmodels.Empty]{
			Error:   true,
			Message: "Both from_time and to_time must be provided",
		})
	}

	if !req.FromTime.Before(req.ToTime) {
		return ctx.JSON(http.StatusBadRequest, httpmodels.Body[httpmodels.Empty]{
			Error:   true,
			Message: "from_time must be before to_time",
		})
	}

	idCombinationUsecase := make([]repository.IDCombination, len(req.IDCombinations))
	for i, val := range req.IDCombinations {
		idCombinationUsecase[i] = repository.IDCombination{
			ID1: val.ID1,
			ID2: val.ID2,
		}
	}
	// Call usecase
	deletedCount, err := usecase.DeleteSensorByIdsAndTime(ctx.Request().Context(), &idCombinationUsecase, req.FromTime, req.ToTime)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, httpmodels.Body[httpmodels.Empty]{
			Error:   true,
			Message: err.Error(),
		})
	}

	body := httpmodels.Body[model.DeleteResponse]{
		Data: model.DeleteResponse{
			DeletedCount: deletedCount.Count,
			Message:      deletedCount.Message,
		},
	}
	return ctx.JSON(http.StatusOK, body)
}

func (s *SensorRouterImpl) UpdateSensorByIds(ctx echo.Context) error {
	return nil
}

func (s *SensorRouterImpl) UpdateSensorByTime(ctx echo.Context) error {
	return nil
}

func (s *SensorRouterImpl) UpdateSensorByIdsAndTime(ctx echo.Context) error {
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
	body := httpmodels.Body[model.SensorPayload]{
		Data: model.SensorPayload{
			ID1:         "10",
			SensorType:  "alksjd",
			Value:       10,
			TimestampMs: time.Now().Unix(),
		},
	}
	return ctx.JSON(http.StatusOK, body)
}

package generator

import (
	"fmt"
	"log"
	"net/http"
	"time"

	httpModels "github.com/Yusufzhafir/worlder-team-assignment/a-service/shared/model"
	"github.com/Yusufzhafir/worlder-team-assignment/a-service/usecase"
	"github.com/labstack/echo/v4"
)

type GeneratorRouter interface {
	Config(ctx echo.Context) error
	Start(ctx echo.Context) error
	Stop(ctx echo.Context) error
	GetStats(ctx echo.Context) error
	SetFrequency(ctx echo.Context) error
	SpamRequests(ctx echo.Context) error
}

type generatorRouterImpl struct {
	usecase *usecase.DataGenerator
}

func NewGeneratorRouter(uscase *usecase.DataGenerator) GeneratorRouter {
	return &generatorRouterImpl{
		usecase: uscase,
	}
}

// Config godoc
// @Summary Configure sensor parameters
// @Description Update sensor configuration including value, type, IDs, and server address
// @Tags generator
// @Accept json
// @Produce json
// @Param config body usecase.GeneratorConfig true "Generator configuration"
// @Success     200 {object} model.Envelope{data=model.Empty} "succesfully updated config"
// @Failure     400 {object} model.Envelope{data=model.Empty} "error=true, message explains"
// @Failure     500 {object} model.Envelope{data=model.Empty}
// @Router /config [post]
func (g *generatorRouterImpl) Config(ctx echo.Context) error {
	config := usecase.GeneratorConfig{}
	if err := ctx.Bind(&config); err != nil {
		return ctx.JSON(http.StatusBadRequest, httpModels.Body[httpModels.Empty]{
			Error:   true,
			Message: fmt.Sprintf("Invalid request body: %v", err),
		})
	}
	(*g.usecase).Config(config)
	return ctx.JSON(http.StatusOK, httpModels.Body[httpModels.Empty]{
		Error:   false,
		Message: "successfully configured message",
	})
}

// GetStats godoc
// @Summary Get generation statistics
// @Description Retrieve current statistics including total sent and failed requests
// @Tags generator
// @Produce json
// @Success 200 {object} model.Envelope{data=map[string]interface{}} "Statistics retrieved successfully"
// @Router /stats [get]
func (g *generatorRouterImpl) GetStats(ctx echo.Context) error {
	result := (*g.usecase).GetDetailedStats()
	log.Default().Printf("output of stats %v", result)
	return ctx.JSON(http.StatusOK, httpModels.Body[interface{}]{
		Data:    result,
		Error:   false,
		Message: "successfully configured message",
	})
}

// swagger:model Empty
type frequencyRequest struct {
	Timeout string `json:"timeout" example:"2s"`
}

// SetFrequency godoc
// @Summary Set generation frequency
// @Description Set the frequency of data generation in requests per second
// @Tags generator
// @Produce json
// @Param       request body frequencyRequest true "frequency request timeout duration"
// @Success 200 {object} model.Envelope{data=frequencyRequest} "Frequency updated successfully"
// @Failure 400 {object} model.Envelope{data=model.Empty} "Invalid frequency value"
// @Router /frequency [post]
func (g *generatorRouterImpl) SetFrequency(ctx echo.Context) error {
	req := frequencyRequest{}
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, httpModels.Body[httpModels.Empty]{
			Error:   true,
			Message: fmt.Sprintf("Invalid request body: %v", err),
		})
	}
	duration, err := time.ParseDuration(req.Timeout)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, httpModels.Body[httpModels.Empty]{
			Error:   true,
			Message: fmt.Sprintf("Invalid request body: %v", err),
		})
	}
	(*g.usecase).SetFrequency(duration)
	return ctx.JSON(http.StatusOK, httpModels.Body[httpModels.Empty]{
		Error:   false,
		Message: "successfully set frequency timeout",
	})
}

// SpamRequests godoc
// @Summary Execute spam requests
// @Description Send multiple concurrent requests to the gRPC service for load testing
// @Tags generator
// @Accept json
// @Produce json
// @Param request body usecase.SpamRequest true "Spam request configuration"
// @Success 200 {object} model.Envelope{data=map[string]interface{}} "Spam requests completed"
// @Failure 400 {object} model.Envelope{data=model.Empty} "Invalid request body"
// @Router /spam [post]
func (g *generatorRouterImpl) SpamRequests(ctx echo.Context) error {
	requestConfig := usecase.SpamRequest{}
	if err := ctx.Bind(&requestConfig); err != nil {
		return ctx.JSON(http.StatusBadRequest, httpModels.Body[httpModels.Empty]{
			Error:   true,
			Message: fmt.Sprintf("Invalid request body: %v", err),
		})
	}
	result := (*g.usecase).SpamRequests(requestConfig)

	return ctx.JSON(http.StatusOK, httpModels.Body[interface{}]{
		Data:    result,
		Error:   false,
		Message: "successfully configured message",
	})
}

// Start godoc
// @Summary Start data generation
// @Description Start continuous data generation to the gRPC service
// @Tags generator
// @Produce json
// @Success 200 {object} model.Envelope{data=model.Empty} "Data generation started successfully"
// @Failure 400 {object} model.Envelope{data=model.Empty} "Generator already running"
// @Router /start [post]
func (g *generatorRouterImpl) Start(ctx echo.Context) error {
	result := (*g.usecase).Start()
	if !result {
		return ctx.JSON(http.StatusBadRequest, httpModels.Body[httpModels.Empty]{
			Error:   true,
			Message: "already running",
		})
	}
	return ctx.JSON(http.StatusOK, httpModels.Body[httpModels.Empty]{
		Error:   false,
		Message: "successfully start message",
	})
}

// Stop godoc
// @Summary Stop data generation
// @Description Stop continuous data generation
// @Tags generator
// @Produce json
// @Success 200 {object} model.Envelope{data=model.Empty} "Data generation stopped successfully"
// @Failure 400 {object} model.Envelope{data=model.Empty} "Generator already stopped"
// @Router /stop [post]
func (g *generatorRouterImpl) Stop(ctx echo.Context) error {
	result := (*g.usecase).Stop()
	if !result {
		return ctx.JSON(http.StatusBadRequest, httpModels.Body[httpModels.Empty]{
			Error:   true,
			Message: "already stop",
		})
	}
	return ctx.JSON(http.StatusOK, httpModels.Body[httpModels.Empty]{
		Error:   false,
		Message: "successfully stop message",
	})
}

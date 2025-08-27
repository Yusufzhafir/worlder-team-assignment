package generator

import (
	"fmt"
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

// Config implements GeneratorRouter.
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

// GetStats implements GeneratorRouter.
func (g *generatorRouterImpl) GetStats(ctx echo.Context) error {
	result1, result2 := (*g.usecase).GetStats()
	return ctx.JSON(http.StatusOK, httpModels.Body[interface{}]{
		Data: map[interface{}]interface{}{
			"total_sent":   result1,
			"total_failed": result2,
		},
		Error:   false,
		Message: "successfully configured message",
	})
}

type frequencyRequest struct {
	Timeout time.Duration `json:"timeout" example:"2s"`
}

// SetFrequency implements GeneratorRouter.
func (g *generatorRouterImpl) SetFrequency(ctx echo.Context) error {
	req := frequencyRequest{}
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, httpModels.Body[httpModels.Empty]{
			Error:   true,
			Message: fmt.Sprintf("Invalid request body: %v", err),
		})
	}
	(*g.usecase).SetFrequency(req.Timeout)
	return ctx.JSON(http.StatusOK, httpModels.Body[httpModels.Empty]{
		Error:   false,
		Message: "successfully set frequency timeout",
	})
}

// SpamRequests implements GeneratorRouter.
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

// Start implements GeneratorRouter.
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

// Stop implements GeneratorRouter.
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

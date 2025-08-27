package router

import (
	"net/http"

	_ "github.com/Yusufzhafir/worlder-team-assignment/a-service/docs"
	"github.com/Yusufzhafir/worlder-team-assignment/a-service/router/generator"
	"github.com/labstack/echo/v4"
	echoSwagger "github.com/swaggo/echo-swagger"
)

type Body[T any] struct {
	Data    T      `json:"data"`
	Error   bool   `json:"error"`
	Message string `json:"message"`
}

func toggleDataSpeed(e echo.Context) error {
	return e.JSON(http.StatusOK, Body[interface{}]{
		Data:    nil,
		Error:   false,
		Message: "successfully toggled speed",
	})
}

func bindGeneratorRoute(group *echo.Group, router generator.GeneratorRouter) {
	group.POST("/config", router.Config)
	group.POST("/start", router.Start)
	group.POST("/stop", router.Stop)
	group.GET("/stats", router.GetStats)
	group.POST("/frequency", router.SetFrequency)
	group.POST("/spam", router.SpamRequests)
}

type BindRouterOpts struct {
	Router generator.GeneratorRouter
	E      *echo.Echo
}

func BindRouter(opts BindRouterOpts) {
	opts.E.GET("/swagger/*", echoSwagger.WrapHandler)
	apiGroup := opts.E.Group("/api/v1")
	bindGeneratorRoute(apiGroup, opts.Router)
}

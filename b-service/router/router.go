package router

import (
	"fmt"

	"github.com/labstack/echo/v4"
)

func BindSensor(e *echo.Group, sensorRouter *SensorRouter) error {
	router := *sensorRouter
	if router == nil {
		return fmt.Errorf("router is empty %v", router)
	}
	e.GET("/sensor", router.GetSensorDataPaginated)
	e.GET("/sensor/id", router.GetSensorDataById)
	e.GET("/sensor/time", router.GetSensorDataByTime)
	e.GET("/sensor/id-time", router.GetSensorDataByIdAndTime)
	e.DELETE("/sensor/delete/id", router.DeleteSensorById)
	e.DELETE("/sensor/delete/time", router.DeleteSensorByTime)
	e.DELETE("/sensor/delete/id-time", router.DeleteSensorByIdAndTime)
	e.PUT("/sensor/update/id", router.UpdateSensorById)
	e.PUT("/sensor/update/time", router.UpdateSensorByTime)
	e.PUT("/sensor/update/id-time", router.UpdateSensorByIdAndTime)
	return nil
}

func BindOthers(e *echo.Group) {
}

type BindRouterOpts struct {
	e            *echo.Echo
	sensorRouter *SensorRouter
}

func BindRouter(opts BindRouterOpts) {
	apiGroup := opts.e.Group("/api/v1")
	BindSensor(apiGroup, opts.sensorRouter)
	BindOthers(apiGroup)
}

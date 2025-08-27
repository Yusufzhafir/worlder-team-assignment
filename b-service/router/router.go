package router

import (
	"fmt"

	sensorRouter "github.com/Yusufzhafir/worlder-team-assignment/b-service/router/sensor"
	"github.com/labstack/echo/v4"
)

func bindSensor(e *echo.Group, sensorRouter *sensorRouter.SensorRouter) error {
	router := *sensorRouter
	if router == nil {
		return fmt.Errorf("router is empty %v", router)
	}
	e.GET("/sensor", router.GetSensorDataPaginated)                      //with no q-param
	e.GET("/sensor/ids", router.GetSensorDataByIds)                      //with no q-param id
	e.GET("/sensor/time", router.GetSensorDataByTime)                    //with no q-param time
	e.GET("/sensor/ids-time", router.GetSensorDataByIdsAndTime)          //with no q-param id and time
	e.DELETE("/sensor/delete/ids", router.DeleteSensorByIds)             //with no q-param id
	e.DELETE("/sensor/delete/time", router.DeleteSensorByTime)           //with no q-param time
	e.DELETE("/sensor/delete/ids-time", router.DeleteSensorByIdsAndTime) //with no q-param id and time
	e.PUT("/sensor/update/ids", router.UpdateSensorByIds)                //with no q-param id
	e.PUT("/sensor/update/time", router.UpdateSensorByTime)              //with no q-param time
	e.PUT("/sensor/update/ids-time", router.UpdateSensorByIdsAndTime)    //with no q-param id and time
	return nil
}

func bindOthers(e *echo.Group) {
}

type BindRouterOpts struct {
	E            *echo.Echo
	SensorRouter *sensorRouter.SensorRouter
}

func BindRouter(opts BindRouterOpts) {
	apiGroup := opts.E.Group("/api/v1")
	bindSensor(apiGroup, opts.SensorRouter)
	bindOthers(apiGroup)
}

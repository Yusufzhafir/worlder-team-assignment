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
	e.GET("/sensor", router.GetSensorDataPaginated)                    //with no q-param
	e.GET("/sensor/ids", router.GetSensorDataById)                     //with no q-param id
	e.GET("/sensor/time", router.GetSensorDataByTime)                  //with no q-param time
	e.GET("/sensor/id-time", router.GetSensorDataByIdAndTime)          //with no q-param id and time
	e.DELETE("/sensor/delete/id", router.DeleteSensorById)             //with no q-param id
	e.DELETE("/sensor/delete/time", router.DeleteSensorByTime)         //with no q-param time
	e.DELETE("/sensor/delete/id-time", router.DeleteSensorByIdAndTime) //with no q-param id and time
	e.PUT("/sensor/update/id", router.UpdateSensorById)                //with no q-param id
	e.PUT("/sensor/update/time", router.UpdateSensorByTime)            //with no q-param time
	e.PUT("/sensor/update/id-time", router.UpdateSensorByIdAndTime)    //with no q-param id and time
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

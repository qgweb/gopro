package router

import (
	"github.com/goweb/gopro/qianzhao/controller"
	"github.com/labstack/echo"
)

func Router(e *echo.Echo) {
	e.Get("/", controller.Index)
}

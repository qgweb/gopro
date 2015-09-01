package controller

import (
	"net/http"

	"github.com/labstack/echo"
)

func Index(ctx *echo.Context) error {
	return ctx.String(http.StatusOK, "index")
}

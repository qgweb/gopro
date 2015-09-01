package main

import (
	"net/http"
	"os"
	"strconv"

	"github.com/astaxie/beego/grace"
	"github.com/labstack/echo"
)

func main() {
	// Setup
	e := echo.New()
	e.Get("/", func(c *echo.Context) error {
		return c.String(http.StatusOK, strconv.Itoa(os.Getpid()))
	})

	grace.ListenAndServe(":1312", e)
}

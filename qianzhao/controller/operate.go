package controller

import (
	"github.com/labstack/echo"
	"net/http"
)

type Operate struct {
}

// 提示
func (this *Operate) SpeedupPrepare(ctx echo.Context) error {
	return ctx.Render(http.StatusOK, "speedup_prepare", "")
}

package controller

import (
	"goclass/convert"
	"log"
	"net/http"

	"github.com/goweb/gopro/qianzhao/common/global"
	"github.com/goweb/gopro/qianzhao/common/session"
	"github.com/goweb/gopro/qianzhao/model"
	"github.com/labstack/echo"
)

type Base struct {
}

func (this *Base) IsLogin(ctx *echo.Context) (bool, error) {
	sess, err := session.GetSession(ctx)
	if err != nil {
		log.Println("获取session失败：", err)
	}

	defer sess.SessionRelease(ctx.Response())

	if _, ok := sess.Get(global.SESSION_USER_INFO).(model.User); !ok {
		return false, ctx.JSON(http.StatusOK, map[string]string{
			"code": convert.ToString(http.StatusMovedPermanently),
			"msg":  global.CONTROLLER_USER_LOGIN_FIRST,
		})
	} else {
		return true, nil
	}
}

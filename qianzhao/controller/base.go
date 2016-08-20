package controller

import (
	"github.com/ngaut/log"
	"net/http"

	"github.com/qgweb/gopro/lib/convert"

	"github.com/labstack/echo"
	"github.com/qgweb/gopro/qianzhao/common/global"
	"github.com/qgweb/gopro/qianzhao/common/session"
	"github.com/qgweb/gopro/qianzhao/model"
	"github.com/labstack/echo/engine/standard"
)

type Base struct {
}

func (this *Base) IsLogin(ctx echo.Context) (bool, error) {
	sess, err := session.GetSession(ctx)
	if err != nil {
		log.Error("获取session失败：", err)
		return false, err
	}

	defer sess.SessionRelease(ctx.Response().(*standard.Response).ResponseWriter)

	if _, ok := sess.Get(global.SESSION_USER_INFO).(model.User); !ok {
		return false, ctx.JSON(http.StatusOK, map[string]string{
			"code": convert.ToString(http.StatusMovedPermanently),
			"msg":  global.CONTROLLER_USER_LOGIN_FIRST,
		})
	} else {
		return true, nil
	}
}

// 获取用户信息
func (this *Base) GetUserInfo(ctx echo.Context) (ui model.User) {
	sess, err := session.GetSession(ctx)
	if err != nil {
		log.Error("获取session失败：", err)
		return
	}

	defer sess.SessionRelease(ctx.Response().(*standard.Response).ResponseWriter)

	if u, ok := sess.Get(global.SESSION_USER_INFO).(model.User); ok {
		return u
	}
	return
}

func (this *Base) SetSess(ctx echo.Context, key string, v interface{}) error {
	sess, err := session.GetSession(ctx)
	if err != nil {
		return err
	}
	defer sess.SessionRelease(ctx.Response().(*standard.Response).ResponseWriter)
	return sess.Set(key, v)
}

func (this *Base) GetSess(ctx echo.Context, key string) (interface{}, error) {
	sess, err := session.GetSession(ctx)
	if err != nil {
		return nil, err
	}
	defer sess.SessionRelease(ctx.Response().(*standard.Response).ResponseWriter)
	return sess.Get(key), nil
}

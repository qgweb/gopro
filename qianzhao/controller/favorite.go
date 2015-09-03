package controller

import (
	"net/http"

	"github.com/goweb/gopro/qianzhao/common/global"
	"github.com/goweb/gopro/qianzhao/model"
	"github.com/labstack/echo"
)

type Favorite struct {
}

// 获取收藏夹
func (this *Favorite) GetFavorite(ctx *echo.Context) error {
	var (
		username = ctx.Form("username")
		fmodel   = model.Favorite{}
		umodel   = model.User{}
	)

	uid := umodel.UserInfo(username).Id
	f := fmodel.GetFavorite(uid)

	if f.Uid != "" {
		return ctx.JSON(http.StatusOK, map[string]string{
			"code": global.CONTROLLER_CODE_SUCCESS,
			"data": f.Content,
		})
	} else {
		return ctx.JSON(http.StatusOK, map[string]string{
			"code": global.CONTROLLER_CODE_ERROR,
			"msg":  global.CONTROLLER_FAVORITE_NOCONTENT,
		})
	}
}

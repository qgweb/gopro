package controller

import (
	"net/http"

	"github.com/goweb/gopro/qianzhao/common/global"
	"github.com/goweb/gopro/qianzhao/model"
	"github.com/labstack/echo"
)

type Favorite struct {
	Base
}

// 获取收藏夹
func (this *Favorite) GetFavorite(ctx *echo.Context) error {
	// 验证是否登录
	if res, err := this.Base.IsLogin(ctx); !res {
		return err
	}

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

// 备份收藏夹
func (this *Favorite) BackupFavorite(ctx *echo.Context) error {
	// 验证是否登录
	if res, err := this.Base.IsLogin(ctx); !res {
		return err
	}

	var (
		username = ctx.Form("username")
		favorite = ctx.Form("favorite")
		fmodel   = model.Favorite{}
		umodel   = model.User{}
	)

	if favorite == "" {
		return ctx.JSON(http.StatusOK, map[string]string{
			"code": global.CONTROLLER_CODE_ERROR,
			"msg":  global.CONTROLLER_FAVORITE_NOUPLOADCONTENT,
		})
	}

	f := model.Favorite{}
	f.Uid = umodel.UserInfo(username).Id
	f.Content = favorite
	fmodel.SaveFavorite(f)

	return ctx.JSON(http.StatusOK, map[string]string{
		"code": global.CONTROLLER_CODE_SUCCESS,
	})
}

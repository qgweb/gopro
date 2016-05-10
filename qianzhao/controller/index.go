package controller

import (
	oredis "github.com/garyburd/redigo/redis"
	"github.com/labstack/echo"
	"github.com/qgweb/gopro/qianzhao/common/redis"
	"github.com/qgweb/gopro/qianzhao/model"
)

type Index struct {
}

//qzbrower-主版本号.次版本号.修订版本号-类型号
func (this *Index) Update(ctx echo.Context) error {
	var (
		version  = ctx.FormValue("version")
		btype    = ctx.FormValue("type")
		mversion = model.Version{}
	)

	if version == "" || btype == "" {
		return ctx.JSON(200, map[string]interface{}{
			"code": "500",
			"msg":  "参数为空",
			"data": "",
		})
	}

	v := mversion.Update(version, btype)
	return ctx.JSON(200, map[string]interface{}{
		"code": "200",
		"msg":  "",
		"data": map[string]interface{}{
			"is_update":    v.IsUpdate,
			"download_url": v.Url,
			"update_page":  v.Update_page,
		},
	})
}

// 浏览器首页控制
func (this *Index) MainPage(ctx echo.Context) error {
	conn := redis.Get()
	defer conn.Close()
	conn.Do("SELECT", "1")
	page, err := oredis.String(conn.Do("GET", "QIANZHAO_PAGE"))
	if err != nil {
		return ctx.JSON(200, map[string]interface{}{
			"code": "500",
			"msg":  "获取首页失败",
			"data": "",
		})
	}
	return ctx.JSON(200, map[string]interface{}{
		"code": "200",
		"msg":  "",
		"data": page,
	})
}

package controller

import (
	"strconv"

	"github.com/labstack/echo"
	"github.com/qgweb/gopro/qianzhao/model"
)

type Index struct {
}

//qianzhao-主版本号.次版本号.修订版本号-类型号-渠道号
func (this *Index) Update(ctx *echo.Context) error {
	var (
		version  = ctx.Form("version")
		btype    = ctx.Form("type")
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
		"data": map[string]string{
			"is_update":    strconv.FormatBool(v.IsUpdate),
			"download_url": v.Url,
			"update_page":  v.Update_page,
		},
	})
}

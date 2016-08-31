package controller

import (
	"encoding/json"
	"github.com/astaxie/beego/httplib"
	oredis "github.com/garyburd/redigo/redis"
	"github.com/labstack/echo"
	"github.com/ngaut/log"
	"github.com/qgweb/gopro/qianzhao/common/config"
	"github.com/qgweb/gopro/qianzhao/common/redis"
	"github.com/qgweb/gopro/qianzhao/model"
)

type Index struct {
}

type YLJData struct {
	Title string `json:"title"`
	Pic   string `json:"pic"`
	Url   string `json:"url"`
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

// 浏览器首页
func (this *Index) Index(ctx echo.Context) error {
	var wm model.Word
	url := config.GetDefault().Key("yljurl").String()
	data, err := httplib.Get(url).Bytes()
	var yljlist = make([]YLJData, 7)
	if err != nil {
		log.Error(err)
	}
	err = json.Unmarshal(data, &yljlist)
	if err != nil {
		log.Error(err)
	}
	// 获取抽奖字谜
	if w, err := wm.Get(); err == nil && w.Id > 0 {
		yljlist[1] = YLJData{
			Pic:   w.Pic,
			Title: w.Title,
			Url:   "/club#three",
		}
	}

	return ctx.Render(200, "index_index", map[string]interface{}{
		"Ylj": yljlist,
	})
}

// 各种统计
package controller

import (
	"fmt"
	"github.com/astaxie/beego/httplib"
	"github.com/bitly/go-simplejson"
	"github.com/labstack/echo"
	"github.com/ngaut/log"
	"github.com/qgweb/gopro/lib/convert"
	"github.com/qgweb/gopro/lib/encrypt"
	"github.com/qgweb/gopro/qianzhao-tcpserver/common/config"
	"github.com/qgweb/gopro/qianzhao/model"
	"net/url"
	"strings"
	"time"
)

type Statistics struct {
}

// 下载统计
func (this *Statistics) Download(ctx *echo.Context) error {
	var (
		q       = ctx.Query("q") //渠道
		v       = ctx.Query("v") //版本号
		t       = ctx.Query("t") //版本 大众版 2,校园版 1
		ip      = strings.Split(ctx.Request().RemoteAddr, ":")[0]
		err     error
		dlmodel = &model.Download{}
	)

	//解码
	if q, err = this.decode(q); err != nil || q == "" {
		return ctx.String(404, "")
	}
	if v, err = this.decode(v); err != nil || v == "" {
		return ctx.String(404, "")
	}
	if t, err = this.decode(t); err != nil || t == "" {
		return ctx.String(404, "")
	}

	go func() {
		dlmodel.Account = this.getAccountByIp(ctx)
		dlmodel.City = this.getCityByIp(ip)
		dlmodel.Date = time.Now().Unix()
		dlmodel.Type = convert.ToInt(t)
		dlmodel.Channel = convert.ToInt(q)
		dlmodel.Version = v

		dlmodel.AddRecord(dlmodel)
	}()

	durl := fmt.Sprintf("%s/qzbrower_%s_%s.exe", config.GetDefault().Key("download").String(), v, t)
	return ctx.Redirect(302, durl)
}

// 日活统计
func (this *Statistics) DayActivity(ctx *echo.Context) error {
	var (
		v       = ctx.Form("v") //版本号
		t       = ctx.Form("t") // 类型
		damodel = &model.DayActivity{}
	)

	if v == "" || t == "" {
		ctx.String(404, "")
	}
	damodel.Count = 1
	damodel.Date = time.Date(time.Now().Year(), time.Now().Month(),
		time.Now().Day(), 0, 0, 0, 0, time.Local).Unix()
	damodel.Type = convert.ToInt(t)
	damodel.Version = v

	damodel.AddRecord(damodel)
	return nil
}

// 侧边栏统计
func (this *Statistics) SideBar(ctx *echo.Context) error {
	var (
		t        = ctx.Form("t")        // 类型
		v        = ctx.Form("v")        //版本号
		favorite = ctx.Form("favorite") //收藏夹
		email    = ctx.Form("email")    //邮箱
		yixin    = ctx.Form("yixin")    //易信
		sbmodel  = &model.SideBar{}
	)

	if t == "" || v == "" {
		ctx.String(404, "")
	}

	if favorite == "" {
		favorite = "0"
	}
	if yixin == "" {
		yixin = "0"
	}
	if email == "" {
		email = "0"
	}

	sbmodel.Date = time.Date(time.Now().Year(), time.Now().Month(),
		time.Now().Day(), 0, 0, 0, 0, time.Local).Unix()
	sbmodel.Type = convert.ToInt(t)
	sbmodel.Favorite = convert.ToInt(favorite)
	sbmodel.Yixin = convert.ToInt(yixin)
	sbmodel.Email = convert.ToInt(email)

	sbmodel.AddRecord(sbmodel)
	return nil
}

// url解码
func (this *Statistics) decode(v string) (string, error) {
	v, err := url.QueryUnescape(v)
	if err != nil {
		return "", err
	}

	return encrypt.DefaultBase64.Decode(v), nil
}

// 通过ip获取账号信息
func (this *Statistics) getAccountByIp(ctx *echo.Context) string {
	bcontroller := &BroadBand{}
	user, err := bcontroller.UserQuery(ctx)
	if err != nil {
		log.Warn(err)
		return ""
	}
	return user.Account
}

// 通过ip获取城市
func (this *Statistics) getCityByIp(ip string) string {
	url := "http://ip.taobao.com/service/getIpInfo.php?ip=" + ip
	if data, err := httplib.Get(url).Bytes(); err == nil {
		js, err := simplejson.NewJson(data)
		if err != nil {
			return ""
		}
		region, _ := js.Get("data").Get("region").String()
		city, _ := js.Get("data").Get("city").String()
		return region + " " + city
	} else {
		return ""
	}
}

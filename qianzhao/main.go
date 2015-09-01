// 千兆浏览器接口
// @authr:郑波
// @date : 2015-09-01
// @version : v0.0.1

package main

import (
	"fmt"

	"github.com/goweb/gopro/qianzhao/common/config"
	"github.com/goweb/gopro/qianzhao/router"

	"github.com/astaxie/beego/grace"
	"github.com/labstack/echo"
	mw "github.com/labstack/echo/middleware"

	"os"
)

const (
	APP_VERSION = "v0.0.1"
)

var ()

// 命令行参数解析
func flagParse() {
	for _, v := range os.Args {
		if v == "-v" {
			fmt.Println("千兆浏览器 [版本号] : ", APP_VERSION)
		}

		if v == "-h" {
			fmt.Println(`
				-v : 版本号
			`)
		}
	}
}

func init() {
	flagParse()
}

func main() {
	var (
		host = config.GetDefault().Key("host").String()
		port = config.GetDefault().Key("port").String()
	)

	e := echo.New()
	e.Use(mw.Recover())
	e.Use(mw.Logger())

	//路由
	router.Router(e)

	grace.ListenAndServe(fmt.Sprintf("%s:%s", host, port), e)
}

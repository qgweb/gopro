// 千兆浏览器接口
// @authr:郑波
// @date : 2015-09-01
// @version : v0.0.1

package main

import (
	"fmt"
	"html/template"
	"io"

	"github.com/goweb/gopro/qianzhao/common/config"
	"github.com/goweb/gopro/qianzhao/router"

	"github.com/astaxie/beego/grace"

	"os"

	_ "github.com/goweb/gopro/qianzhao/common/redis"
	_ "github.com/goweb/gopro/qianzhao/common/session"
	_ "github.com/goweb/gopro/qianzhao/model"
	"github.com/labstack/echo"
	mw "github.com/labstack/echo/middleware"
)

const (
	APP_VERSION = "v0.0.1"
)

type Template struct {
	templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

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

	t := &Template{
		templates: template.Must(template.ParseGlob("views/*/*.html")),
	}

	e := echo.New()
	e.Use(mw.Recover())
	e.Use(mw.Logger())

	e.Favicon("public/favicon.ico")
	e.Static("/", "public/")
	e.Index("public/index.html")
	//路由
	router.Router(e)

	e.SetRenderer(t)

	grace.ListenAndServe(fmt.Sprintf("%s:%s", host, port), e)
}

// 设置错误处理
//e.SetHTTPErrorHandler(func(err error, c *echo.Context) {
//	fmt.Println(err)
//	http.Error(c.Response(), "fuck", 404)
//})

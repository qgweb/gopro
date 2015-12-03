// 千兆浏览器接口
// @authr:郑波
// @date : 2015-09-01
// @version : v0.0.1

package main

import (
	"fmt"
	"html/template"
	"io"
	"os"

	"github.com/labstack/echo"
	mw "github.com/labstack/echo/middleware"
	"github.com/ngaut/log"
	"github.com/qgweb/gopro/qianzhao/common/config"
	//_ "github.com/qgweb/gopro/qianzhao/common/redis"
	_ "github.com/qgweb/gopro/qianzhao/common/session"
	_ "github.com/qgweb/gopro/qianzhao/model"
	"github.com/qgweb/gopro/qianzhao/router"
)

const (
	APP_VERSION = "v0.0.2"
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
			fmt.Println(`-v : 版本号`)
		}
	}
}

// 日志类初始化
func loginit() {
	logFile := config.GetDefault().Key("log").String()
	if logFile == "" {
		log.Fatal("日志文件不存在")
	}

	log.SetHighlighting(false)
	log.SetOutputByName(logFile)
	log.SetRotateByDay()
}

func init() {
	flagParse()
	loginit()
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
	e.Run(fmt.Sprintf("%s:%s", host, port))
}

// 设置错误处理
//e.SetHTTPErrorHandler(func(err error, c *echo.Context) {
//	fmt.Println(err)
//	http.Error(c.Response(), "fuck", 404)
//})

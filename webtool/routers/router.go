package routers

import (
	"github.com/astaxie/beego"
	"github.com/qgweb/gopro/webtool/controllers"
)

func init() {
	beego.AutoRouter(&controllers.MainController{})
}

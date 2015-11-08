package routers

import (
	"github.com/qgweb/gopro/webtool/controllers"
	"github.com/astaxie/beego"
)

func init() {
    beego.Router("/", &controllers.MainController{})
}

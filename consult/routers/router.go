package routers

import (
	"github.com/qgweb/gopro/consult/controllers"
	"github.com/astaxie/beego"
)

func init() {
	beego.Router("/", &controllers.MainController{})
	beego.Router("/submit", &controllers.MainController{}, "post:Submit")
	beego.Router("/list", &controllers.MainController{}, "get:List")
	beego.Router("/sts", &controllers.MainController{}, "get:StatsRec")
}

package routers

import (
	"github.com/astaxie/beego"
	"github.com/qgweb/gopro/consult/controllers"
)

func init() {
	beego.Router("/", &controllers.MainController{})
	beego.Router("/submit", &controllers.MainController{}, "post:Submit")
	beego.Router("/list", &controllers.MainController{}, "get:List")
	beego.Router("/dy",&(controllers.MainController{}),"get,post:Diaoyan")
	beego.Router("/sts", &controllers.MainController{}, "get:StatsRec")
}

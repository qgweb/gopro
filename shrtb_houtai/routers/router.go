package routers

import (
	"github.com/qgweb/gopro/shrtb_houtai/controllers"
	"github.com/astaxie/beego"
)

func init() {
	beego.Router("/login", &controllers.MainController{}, "*:Login")
	beego.Router("/list", &controllers.MainController{}, "*:List")
    beego.Router("/", &controllers.MainController{})
}

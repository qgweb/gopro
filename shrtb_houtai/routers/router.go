package routers

import (
	"github.com/astaxie/beego"
	"github.com/qgweb/gopro/shrtb_houtai/controllers"
)

func init() {
	beego.Router("/login", &controllers.MainController{}, "*:Login")
	beego.Router("/logout", &controllers.MainController{}, "*:Logout")
	beego.Router("/list", &controllers.MainController{}, "*:List")
	beego.Router("/", &controllers.MainController{})
	beego.AutoRouter(&controllers.OrderController{})
	beego.AutoRouter(&controllers.ReportController{})
}

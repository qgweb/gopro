package routers

import (
	"github.com/qgweb/gopro/cprotj/controllers"
	"github.com/astaxie/beego"
)

func init() {
    beego.Router("/cj", &controllers.MainController{}, "get:CookieMatch")
}

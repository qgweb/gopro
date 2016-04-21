package routers

import (
	"github.com/astaxie/beego"
	"github.com/qgweb/gopro/cprotj/controllers"
)

func init() {
	beego.Router("/cj", &controllers.MainController{}, "get:CookieMatch")
	beego.Router("/rf", &controllers.MainController{}, "get:Reffer")
	beego.Router("/if", &controllers.MainController{}, "get:Iframe")
	beego.Router("/da", &controllers.MainController{}, "get:RecordADUA")
}

package main

import (
	"goclass/convert"

	"github.com/astaxie/beego"
	_ "github.com/qgweb/gopro/shrtb_houtai/routers"
	"github.com/qgweb/new/lib/timestamp"
)

func TimeParse(val interface{}) string {
	return timestamp.GetUnixFormat(convert.ToString(val))
}

func main() {
	//beego.BConfig.WebConfig.AutoRender = false
	beego.AddFuncMap("unix", TimeParse)
	beego.Run()
}

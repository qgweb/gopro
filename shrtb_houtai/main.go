package main

import (
	"goclass/convert"

	"github.com/astaxie/beego"
	"github.com/qgweb/gopro/shrtb_houtai/models"
	_ "github.com/qgweb/gopro/shrtb_houtai/routers"
	"github.com/qgweb/new/lib/timestamp"
)

func TimeParse(val interface{}) string {
	return timestamp.GetUnixFormat(convert.ToString(val))
}

func cron() {
	var r = &models.Report{}
	var o = &models.Order{}
	o.Putin()
	go r.LoopPvStats()

}
func main() {
	//beego.BConfig.WebConfig.AutoRender = false
	cron()
	beego.AddFuncMap("unix", TimeParse)
	beego.Run()
}

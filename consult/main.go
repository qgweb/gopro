package main

import (
	"github.com/astaxie/beego"
	_ "github.com/qgweb/gopro/consult/routers"
	"github.com/qgweb/new/lib/convert"
	"github.com/qgweb/new/lib/timestamp"
)

func TimeParse(val interface{}) string {
	return timestamp.GetUnixFormat(convert.ToString(val))
}

func main() {
	beego.AddFuncMap("unix", TimeParse)
	beego.Run()
}

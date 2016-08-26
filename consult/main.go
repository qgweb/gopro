package main

import (
	_ "github.com/qgweb/gopro/consult/routers"
	"github.com/astaxie/beego"
	"github.com/qgweb/new/lib/timestamp"
	"github.com/qgweb/new/lib/convert"
)

func TimeParse(val interface{}) string {
	return timestamp.GetUnixFormat(convert.ToString(val))
}

func main() {
	beego.AddFuncMap("unix", TimeParse)
	beego.Run()
}


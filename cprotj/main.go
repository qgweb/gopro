package main

import (
	_ "github.com/qgweb/gopro/cprotj/routers"
	_ "github.com/qgweb/gopro/cprotj/common/mongo"
	_"github.com/qgweb/gopro/cprotj/common/xredis"
	"github.com/astaxie/beego"
)

func main() {
	beego.SetLogFuncCall(true)
	beego.Run()
}


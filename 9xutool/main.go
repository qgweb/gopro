package main

import (
	"github.com/codegangsta/cli"
	"github.com/qgweb/gopro/9xutool/model"
	"github.com/qgweb/gopro/9xutool/model/blackad"
	"github.com/qgweb/gopro/9xutool/model/visitor"
	"github.com/qgweb/gopro/9xutool/model/lonlat"
	"github.com/qgweb/gopro/9xutool/model/putin"
	"github.com/qgweb/gopro/9xutool/model/domain"
	"os"
)

func main() {
	app := cli.NewApp()
	app.Name = "9xutool"
	app.Usage = "九旭小程序集合封装工具"
	app.Version = "0.0.1"

	app.Commands = []cli.Command{
		model.NewUserTraceCli(),
		model.NewShopTraceCli(),
		model.NewURLTraceCli(),
		model.NewTAGSTraceCli(),
		lonlat.NewTagsNumberCli(),
		lonlat.NewUserCdCli(),
		lonlat.NewTagsCdCli(),
		lonlat.NewLonLatCli(),
		putin.NewZheJiangPutCli(),
		putin.NewJiangSuPutCli(),
		blackad.NewBlackMenuCli(),
		visitor.NewZhejiangVisitorCli(),
		domain.NewURLTraceCli(),
	}

	app.Run(os.Args)
}

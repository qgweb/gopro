package main

import (
	"os"

	"github.com/codegangsta/cli"
	"github.com/qgweb/gopro/9xutool/model"
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
	}

	app.Run(os.Args)
}

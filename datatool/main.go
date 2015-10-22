package main

import (
	"os"

	"github.com/codegangsta/cli"
	"github.com/qgweb/gopro/datatool/model/adcate"
)

func main() {
	app := cli.NewApp()
	app.Name = "datatool"
	app.Usage = "九旭小程序集合封装工具"
	app.Version = "0.0.1"

	app.Commands = []cli.Command{
		adcate.NewAdCate(),
	}

	app.Run(os.Args)
}

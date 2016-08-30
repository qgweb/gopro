package main

import (
	"github.com/codegangsta/cli"
	zjmiddle "github.com/qgweb/new/xcrontab/model/zhejiang/middle"
	"os"
)

func main() {
	app := cli.NewApp()
	app.Name = "xcrontab"
	app.Usage = "九旭任务计划"
	app.Version = "0.0.1"

	app.Commands = []cli.Command{
		zjmiddle.CliDomain(),
		zjmiddle.CliUserTrack(),
	}

	app.Run(os.Args)
}

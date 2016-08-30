package config

import (
	"github.com/ngaut/log"
	"github.com/qgweb/new/lib/common"
	"github.com/qgweb/new/lib/config"
)

var (
	confFile config.ConfigContainer
	err      error
)

func init() {
	initConn()
}

func initConn() {
	fileName := common.GetBasePath() + "/conf/conf.ini"
	confFile, err = config.NewConfig("ini", fileName)
	if err != nil {
		log.Fatal(err)
		return
	}
}

func GetConf() config.ConfigContainer {
	if confFile == nil {
		initConn()
	}
	return confFile
}

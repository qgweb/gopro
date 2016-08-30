package common

import (
	"github.com/ngaut/log"
	"github.com/qgweb/go-hbase"
	"github.com/qgweb/new/lib/config"
	"github.com/qgweb/new/lib/mongodb"
	"github.com/qgweb/new/lib/rediscache"
)

var (
	CommonHbase     hbase.HBaseClient      //数据源hbase
	CommonDataMongo *mongodb.Mongodb       //跑数据mongo
	CommonPutReids  *rediscache.MemCache   //投放redis
	appini          config.ConfigContainer //配置文件
	err             error
)

func init() {
	initHbase()
	initDataMongo()
}

func initConfig() {
	var err error
	appini, err = GetConfObj(GetConfigPath())
	if err != nil {
		log.Fatal(err)
	}
}

func initHbase() {
	if CommonHbase, err = GetHbaseObj(GetConfigPath(), "hbase"); err != nil {
		log.Fatal(err)
	}
}

func initDataMongo() {
	if CommonDataMongo, err = GetMongoObj(GetConfigPath(), "data-mongo"); err != nil {
		log.Fatal(err)
	}
}

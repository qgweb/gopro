package models

import (
	"encoding/json"
	"os"
	"strings"

	"github.com/astaxie/beego"
	"github.com/qgweb/new/lib/mongodb"
	"github.com/qgweb/new/lib/rediscache"
)

var (
	putDb *rediscache.MemCache
	mdb   *mongodb.Mongodb
)

func initDataMongo() {
	var (
		host = beego.AppConfig.String("mongo::host")
		port = beego.AppConfig.String("mongo::port")
		db   = beego.AppConfig.String("mongo::db")
		conf = mongodb.MongodbConf{Host: host, Port: port, Db: db}
		err  error
	)
	mdb, err = mongodb.NewMongodb(conf)
	if err != nil {
		beego.Error(err)
		os.Exit(-2)
	}
}

func initPutDb() {
	var (
		hosts = strings.Split(beego.AppConfig.String("redis::host"), ":")
		db    = beego.AppConfig.String("redis::db")
		err   error
	)
	if putDb == nil {
		conf := rediscache.MemConfig{}
		conf.Host = hosts[0]
		conf.Port = hosts[1]
		putDb, err = rediscache.New(conf)
		if err != nil {
			beego.Error(err)
			os.Exit(-2)
		}
		putDb.SelectDb(db)
	}
}

func init() {
	initDataMongo()
	initPutDb()
}

func Parse(info map[string]interface{}, obj interface{}) interface{} {
	if bs, err := json.Marshal(&info); err == nil {
		if err := json.Unmarshal(bs, obj); err == nil {
			return obj
		}
	}
	return nil
}

func Uparse(info interface{}) map[string]interface{} {
	if bs, err := json.Marshal(&info); err == nil {
		var o map[string]interface{}
		if err := json.Unmarshal(bs, &o); err == nil {
			return o
		}
	}
	return nil
}

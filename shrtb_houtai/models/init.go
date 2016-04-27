package models

import (
	"github.com/astaxie/beego"
	"github.com/qgweb/new/lib/rediscache"
	"github.com/qgweb/gossdb"
	"os"
	"strings"
	"github.com/qgweb/new/lib/convert"
)

var (
	dataDb *gossdb.Connectors
	putDb  *rediscache.MemCache
)

func initDataDb() {
	var (
		hosts = strings.Split(beego.AppConfig.String("ssdb::host"), ":")
		err       error
	)

	dataDb, err = gossdb.NewPool(&gossdb.Config{
		Host:             hosts[0],
		Port:             convert.ToInt(hosts[1]),
		MinPoolSize:      5,
		MaxPoolSize:      50,
		AcquireIncrement: 5,
		GetClientTimeout: 10,
		MaxWaitSize:      1000,
		MaxIdleTime:      1,
		HealthSecond:     2,
	})

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
	initDataDb()
	initPutDb()
}

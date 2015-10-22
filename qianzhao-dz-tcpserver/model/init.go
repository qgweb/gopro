package model

import (
	"fmt"
	"log"

	"github.com/qgweb/gopro/lib/orm"
	"github.com/qgweb/gopro/qianzhao-dz-tcpserver/common/config"
)

var (
	myorm *orm.QGORM
)

func init() {
	var (
		user    = config.GetDB().Key("user").String()
		pwd     = config.GetDB().Key("pwd").String()
		port    = config.GetDB().Key("port").String()
		host    = config.GetDB().Key("host").String()
		db      = config.GetDB().Key("db").String()
		charset = config.GetDB().Key("charset").String()
	)

	myorm = orm.NewORM()

	err := myorm.Open(fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s",
		user, pwd, host, port, db, charset))

	if err != nil {
		log.Fatalln("连接数据库失败：", err)
	}

	myorm.SetMaxIdleConns(50)
	myorm.SetMaxOpenConns(100)
}

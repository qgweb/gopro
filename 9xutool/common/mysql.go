package common

import (
	"fmt"
	"github.com/qgweb/gopro/lib/orm"
	"time"

	"github.com/ngaut/log"
	"gopkg.in/mgo.v2"
)

type MysqlConfig struct {
	UserName string
	UserPwd  string
	Host     string
	Port     string
	DBName   string
}

type MysqlPool struct {
	db   *mgo.Session
	conf *MysqlConfig
}

func NewMysqlPool(conf *MgoConfig) *MgoPool {
	return &MgoPool{conf: conf}
}

func (this *QGORM) Get() *mgo.Session {
	this.mysql = orm.NewORM()
	err := this.mysql.Open(fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s",
		user, pwd, host, port, db, charset))

	if err != nil {
		log.Fatal("连接数据库失败：", err)
	}

	myorm.SetMaxIdleConns(50)
	myorm.SetMaxOpenConns(100)
}

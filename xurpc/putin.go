package main

import (
	"errors"
	"fmt"
	"github.com/ngaut/log"
	"github.com/qgweb/new/lib/config"
	"goclass/orm"
)

type Putin struct {
}
type Order interface {
	Handler()
	Check()
}
type Advert struct {
	id     string
	status string
}
type Strategy struct {
	id     string
	status string
}

var (
	mysqlsession *orm.QGORM
	// mongosession *mongodb.Mongodb
)

func (this Putin) StatusHandler(id, category, status string) []byte {
	var order Order
	if id == "" || category == "" || status == "" {
		return jsonReturn("", errors.New("参数不能为空"))
	}
	switch category {
	case "advert":
		order = &Advert{id: id, status: status}
	case "strategy":
		order = &Strategy{id: id, status: status}
	default:
		jsonReturn("", errors.New("类型错误"))
	}

	getMysqlSession()
	// getMongoSession()
	handleData(order)
}

func handleData(order Order) {
	order.Check()
	order.Handler()
}

func (this *Advert) Check() {
	mysqlsession.BSQL().Select("*").From("nxu_advert").
}

func (this *Advert) Handler() {
	return
}

func (this *Strategy) Check() {
	return
}

func (this *Strategy) Handler() {
	return
}

func getMysqlSession() {
	iniFile, err := config.NewConfig("ini", *conf)
	if err != nil {
		log.Fatal("open configfile error:", err)
	}
	var (
		host    = iniFile.String("mysql-9xu::host")
		port    = iniFile.String("mysql-9xu::port")
		user    = iniFile.String("mysql-9xu::user")
		pwd     = iniFile.String("mysql-9xu::pwd")
		db      = iniFile.String("mysql-9xu::db")
		charset = "utf8"
	)
	err = mysqlsession.Open(fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s", user, pwd, host, port, db, charset))
	if err != nil {
		log.Fatal("db connect error:", err)
	}
	mysqlsession.SetMaxIdleConns(50)
	mysqlsession.SetMaxOpenConns(50)
}

// func getMongoSession() {
// 	var (
// 		host = iniFile.String("mysql-9xu::host")
// 		port = iniFile.String("mysql-9xu::port")
// 		user = iniFile.String("mysql-9xu::user")
// 		pwd  = iniFile.String("mysql-9xu::pwd")
// 		db   = iniFile.String("mysql-9xu::db")
// 	)
// 	mgoConfig := mongodb.MongodbConf{host, port, user, pwd, db}
// 	mongodb, err = mongodb.NewMongodb(mgoConfig)
// 	if err != nil {
// 		log.Fatal("mongodb connect error:", err)
// 	}
// }

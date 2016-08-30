package db

import (
	"fmt"
	"github.com/astaxie/beego/orm"
	_ "github.com/go-sql-driver/mysql"
	"github.com/qgweb/new/xrpc/config"
)

var (
	mysqlConn orm.Ormer
)

func init() {
	initMysqlConn()
}

func initMysqlConn() {
	var (
		host = config.GetConf().String("mysql::host")
		port = config.GetConf().String("mysql::port")
		db   = config.GetConf().String("mysql::db")
		user = config.GetConf().String("mysql::user")
		pwd  = config.GetConf().String("mysql::pwd")
	)
	orm.RegisterDataBase("default", "mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8", user, pwd, host, port, db), 100)
	mysqlConn = orm.NewOrm()
}
func GetMysqlConn() orm.Ormer {
	if mysqlConn == nil {
		initMysqlConn()
	}
	return mysqlConn
}

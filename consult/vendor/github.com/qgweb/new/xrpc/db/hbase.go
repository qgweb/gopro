package db

import (
	"github.com/ngaut/log"
	"github.com/qgweb/go-hbase"
	"github.com/qgweb/new/xrpc/config"
	"strings"
)

var (
	hbaseConn hbase.HBaseClient
	err       error
)

func init() {
	initHbaseConn()
}

func initHbaseConn() {
	var (
		host = config.GetConf().String("hbase::host")
		port = config.GetConf().String("hbase::port")
	)

	hosts := strings.Split(host, ",")
	for k, _ := range hosts {
		hosts[k] += ":" + port
	}

	hbaseConn, err = hbase.NewClient(hosts, "/hbase")
	if err != nil {
		log.Error(err)
	}
}

func GetHbaseConn() hbase.HBaseClient {
	if hbaseConn == nil {
		initHbaseConn()
	}
	return hbaseConn
}

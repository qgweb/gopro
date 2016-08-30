package db

import (
	"github.com/bradfitz/gomemcache/memcache"
	"github.com/qgweb/new/xrpc/config"
)

var (
	mClient *memcache.Client
)

func init() {
	initMemcacheConn()
}

func initMemcacheConn() {
	var (
		host = config.GetConf().String("memcache::host")
		port = config.GetConf().String("memcache::port")
	)

	mClient = memcache.New(host + ":" + port)
}

func GetMemcacheConn() *memcache.Client {
	if mClient == nil {
		initMemcacheConn()
	}
	return mClient
}

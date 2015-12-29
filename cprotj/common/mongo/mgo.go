package mongo

import (
	"fmt"
	"github.com/astaxie/beego"
	"gopkg.in/mgo.v2"
	"os"
	"sync"
	"time"
	"strings"
)

var (
	mongodb   *mgo.Session
	mgoconfig *MongoConfig
	mux       sync.RWMutex
)

type MongoConfig struct {
	Host string
	Port string
	Db   string
	User string
	Pwd  string
	At   string //@
}

func getMongoConfig() *MongoConfig {
	mc := &MongoConfig{}
	config, err := beego.AppConfig.GetSection("mongo")
	if err != nil {
		beego.Emergency(err)
		os.Exit(-2)
	}
	if v, ok := config["host"]; ok {
		mc.Host = v
	}
	if v, ok := config["port"]; ok {
		mc.Port = v
	}
	if v, ok := config["db"]; ok {
		mc.Db = v
	}
	if v, ok := config["user"]; ok {
		mc.User = v
	}
	if v, ok := config["pwd"]; ok {
		mc.Pwd = v
	}
	if mc.User != "" && mc.Pwd != "" {
		mc.At = "@"
	}
	return mc
}

func init() {
	mgoconfig = getMongoConfig()
}

func LinkMongo() (*mgo.Session, error) {
	var err error
	mux.Lock()
	defer mux.Unlock()
	if mongodb == nil {
		url := fmt.Sprintf("%s:%s", mgoconfig.User, mgoconfig.Pwd) + mgoconfig.At +
			fmt.Sprintf("%s:%s/%s", mgoconfig.Host, mgoconfig.Port, mgoconfig.Db)
		url = strings.TrimPrefix(url, ":")
		mongodb, err = mgo.DialWithTimeout(url, time.Second*10)
		if err != nil {
			return nil, err
		}
	}

	err = mongodb.Ping()
	if err != nil {
		return nil, err
	}
	return mongodb.Clone(), nil
}

func GetConfig() *MongoConfig{
	mux.RLock()
	defer mux.RUnlock()
	return mgoconfig
}

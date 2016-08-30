package common

import (
	"github.com/juju/errors"
	"github.com/qgweb/go-hbase"
	"github.com/qgweb/new/lib/config"
	"github.com/qgweb/new/lib/mongodb"
	"github.com/qgweb/new/lib/rediscache"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// 获取程序执行目录
func GetBasePath() string {
	file, _ := exec.LookPath(os.Args[0])
	path, _ := filepath.Abs(file)
	return filepath.Dir(path)
}

// 获取配置文件对象
func GetConfigPath() string {
	return GetBasePath() + "/conf/conf.ini"
}

// 获取配置文件对象
func GetConfObj(iniPath string) (config.ConfigContainer, error) {
	return config.NewConfig("ini", iniPath)
}

// 获取mongodb对象
func GetMongoObj(iniPath string, section string) (*mongodb.Mongodb, error) {
	confObj, err := GetConfObj(iniPath)
	if err != nil {
		return nil, errors.Trace(err)
	}
	conf := mongodb.MongodbConf{}
	conf.Host = confObj.String(section + "::" + "host")
	conf.Db = confObj.String(section + "::" + "db")
	conf.Port = confObj.String(section + "::" + "port")
	conf.UName = confObj.String(section + "::" + "user")
	conf.Upwd = confObj.String(section + "::" + "pwd")
	return mongodb.NewMongodb(conf)
}

// 获取redis对象
func GetRedisObj(iniPath string, section string, db string) ([]*rediscache.MemCache, error) {
	confObj, err := GetConfObj(iniPath)
	if err != nil {
		return nil, errors.Trace(err)
	}
	conf := rediscache.MemConfig{}
	conf.Host = confObj.String(section + "::" + "host")
	conf.Port = confObj.String(section + "::" + "port")
	conf.Db = db
	return rediscache.NewMul(conf, 2)
}

// 获取hbase对象
func GetHbaseObj(iniPath string, section string) (hbase.HBaseClient, error) {
	confObj, err := GetConfObj(iniPath)
	if err != nil {
		return nil, errors.Trace(err)
	}
	host := confObj.String(section + "::" + "host")
	port := confObj.String(section + "::" + "port")
	hosts := strings.Split(host, ",")
	for k, _ := range hosts {
		hosts[k] += ":" + port
	}
	return hbase.NewClient(hosts, "/hbase")
}

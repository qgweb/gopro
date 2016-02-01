package common

import (
	"github.com/juju/errors"
	"github.com/qgweb/gopro/lib/convert"
	"github.com/qgweb/new/lib/config"
	"github.com/qgweb/new/lib/mongodb"
	"github.com/qgweb/new/lib/rediscache"
	"github.com/qgweb/go-hbase"
	"os"
	"os/exec"
	"path/filepath"
	"time"
	"strings"
)

// 获取程序执行目录
func GetBasePath() string {
	file, _ := exec.LookPath(os.Args[0])
	path, _ := filepath.Abs(file)
	return filepath.Dir(path)
}

/**
 * 获取任意一天整点时间戳
 */
func GetDayTimestamp(day int) string {
	d := time.Now().AddDate(0, 0, day).Format("20060102")
	a, _ := time.ParseInLocation("20060102", d, time.Local)
	return convert.ToString(a.Unix())
}

/**
 * 获取任意一天整点时间戳
 */
func GetDayTimestampFormat(day int, format string) string {
	d := time.Now().AddDate(0, 0, day).Format(format)
	a, _ := time.ParseInLocation(format, d, time.Local)
	return convert.ToString(a.Unix())
}

func GetHourTimestamp(hour int) string {
	d := time.Now().Add(time.Hour * time.Duration(hour)).Format("2006010215")
	a, _ := time.ParseInLocation("2006010215", d, time.Local)
	return convert.ToString(a.Unix())
}

//获取时间字符串
func GetDay(day int) (tf string) {
	t := time.Now()
	if day != 0 {
		t = t.AddDate(0, 0, day)
	}
	tf = t.Format("20060102")
	return
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
func GetRedisObj(iniPath string, section string) (*rediscache.MemCache, error) {
	confObj, err := GetConfObj(iniPath)
	if err != nil {
		return nil, errors.Trace(err)
	}
	conf := rediscache.MemConfig{}
	conf.Host = confObj.String(section + "::" + "host")
	conf.Port = confObj.String(section + "::" + "port")
	conf.Db = confObj.String(section + "::" + "db")
	return rediscache.New(conf)
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
	for k,_ := range hosts {
		hosts[k] += ":" + port
	}
	return hbase.NewClient(hosts, "/hbase")
}

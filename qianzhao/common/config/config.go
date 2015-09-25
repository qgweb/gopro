package config

import (
	"github.com/ngaut/log"
	"io/ioutil"

	"github.com/qgweb/gopro/qianzhao/common/function"
	"gopkg.in/ini.v1"
)

var (
	Config *ini.File
	err    error
)

func init() {
	fname := function.GetBasePath() + "/conf/app.ini"
	data, err := ioutil.ReadFile(fname)
	if err != nil {
		log.Fatal("打开配置文件出错，错误信息为：", err)
	}

	Config, err = ini.Load(data)
	if err != nil {
		log.Fatal("加载配置文件出错，错误信息为：", err)
	}
}

// 获取默认配置
func GetDefault() *ini.Section {
	return Config.Section("default")
}

// 获取数据库配置
func GetDB() *ini.Section {
	return Config.Section("mysql")
}

// 获取redis配置
func GetRedis() *ini.Section {
	return Config.Section("redis")
}

// 活取接口配置
func GetInterface() *ini.Section {
	return Config.Section("interface")
}

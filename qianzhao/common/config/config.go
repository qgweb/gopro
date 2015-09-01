package config

import (
	"io/ioutil"
	"log"

	"github.com/goweb/gopro/qianzhao/common/function"
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
		log.Fatalln("打开配置文件出错，错误信息为：", err)
	}

	Config, err = ini.Load(data)
	if err != nil {
		log.Fatalln("加载配置文件出错，错误信息为：", err)
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

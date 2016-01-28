package common

import (
	"flag"
	"github.com/qgweb/new/lib/config"
	"gopkg.in/ini.v1"
	"io/ioutil"
	"log"
)

var (
	conf = flag.String("conf", "db.ini", "配置文件")
)

func init() {
	flag.Parse()
}

func GetIniObject() (*ini.File, error) {
	source, err := ioutil.ReadFile(*conf)
	if err != nil {
		log.Fatalln(err)
	}
	iniFile, err := ini.Load(source)
	return iniFile, err
}

func GetBeegoIniObject() (config.ConfigContainer, error) {
	return config.NewConfig("ini", *conf)
}

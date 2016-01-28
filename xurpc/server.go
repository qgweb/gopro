package main

import (
	"fmt"
	"github.com/hprose/hprose-go/hprose"
	"github.com/wangbin/jiebago/analyse"
	"gopkg.in/ini.v1"
	"log"
	"net/http"
	"github.com/qgweb/gopro/xurpc/putin"
	_"github.com/qgweb/gopro/xurpc/common"
	"github.com/qgweb/gopro/xurpc/common"
)

var (
	IniFile *ini.File
	seg     analyse.TagExtracter //结巴分词
)

func init() {
	var err error
	IniFile,err = common.GetIniObject()
	if err !=nil{
		log.Fatalln(err)
		return
	}
	loadDic()
}

func loadDic() {
	err := seg.LoadDictionary("./dictionary/dict.txt")
	if err != nil {
		log.Fatal("打开字典文件错误")
	}
	err = seg.LoadIdf("./dictionary/idf.txt")
	if err != nil {
		log.Fatal("打开逆向字典文件错误")
	}
}

func main() {
	host := IniFile.Section("pro").Key("host").String()
	port := IniFile.Section("pro").Key("port").String()

	if host == "" || port == "" {
		log.Fatalln("host或port不能为空")
	}

	service := hprose.NewHttpService()

	//服务注册入口
	//service.AddFunction("hello", hello)dd
	service.AddMethods(Taotag{})
	service.AddMethods(TaoShop{})
	service.AddMethods(TaoCatData{})
	service.AddMethods(UserCookieData{})
	service.AddMethods(SearchWord{})
	service.AddMethods(putin.Putin{})

	//服务开启
	http.ListenAndServe(fmt.Sprintf("%s:%s", host, port), service)
}

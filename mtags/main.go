package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"runtime"

	"gopkg.in/ini.v1"

	"github.com/bitly/go-nsq"
)

var (
	conf    = flag.String("conf", "conf.ini", "配置文件")
	iniFile *ini.File
)

func init() {
	flag.Parse()
	confData, err := ioutil.ReadFile(*conf)
	if err != nil {
		log.Fatalln("读取配置文件出错,错误信息为:", err)
	}

	iniFile, err = ini.Load(confData)
	if err != nil {
		log.Fatalln("读取配置文件内容出错,错误信息为:", err)
	}
}

type TailHandler struct {
}

func (th *TailHandler) HandleMessage(m *nsq.Message) error {
	data := &CombinationData{}
	err := json.Unmarshal(m.Body, data)
	if err != nil {
		log.Println("数据解析出错,错误信息为:", err)
		return err
	}

	//分配
	dispath(data)

	return nil
}

func bootstrap() {
	var (
		host  = iniFile.Section("nsq").Key("host").String()
		nport = iniFile.Section("nsq").Key("nport").String()
		lport = iniFile.Section("nsq").Key("lport").String()
		key   = iniFile.Section("nsq").Key("key").String()
	)
	cus, err := nsq.NewConsumer(key, "goods", nsq.NewConfig())
	if err != nil {
		log.Fatalln("连接nsq失败,错误信息为:", err)
	}

	cus.AddHandler(&TailHandler{})
	fmt.Println(cus.ConnectToNSQD(fmt.Sprintf("%s:%s", host, nport)))
	fmt.Println(cus.ConnectToNSQLookupd(fmt.Sprintf("%s:%s", host, lport)))

	for {
		select {
		case <-cus.StopChan:
			return
		}
	}
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	bootstrap()
}

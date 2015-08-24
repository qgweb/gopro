package main

import (
	"flag"
	"fmt"
	"github.com/bitly/go-nsq"
	"log"
	"net/http"
	"strings"
)

var (
	host    = flag.String("host", "127.0.0.1", "http绑定地址")
	port    = flag.String("port", "1219", "http绑定端口")
	nsqhost = flag.String("nsq_host", "127.0.0.1", "nsq地址")
	nsqport = flag.String("nsq_port", "4150", "nsq端口")
	auth    = flag.String("auth", "", "验证密码")
	nsqpro  *nsq.Producer
	err     error
)

func init() {
	flag.Parse()
	initNsq()
}

func initNsq() {
	if *nsqhost == "" || *nsqport == "" {
		log.Fatalln("nsq地址或者端口为空")
		return
	}

	nsqpro, err = nsq.NewProducer(fmt.Sprintf("%s:%s", *nsqhost, *nsqport), nsq.NewConfig())
	if err != nil {
		log.Fatalln("初始化nsq producer 失败")
		return
	}
}

func main() {
	http.HandleFunc("/", Proxy)
	http.ListenAndServe(fmt.Sprintf("%s:%s", *host, *port), nil)
}

func Proxy(w http.ResponseWriter, r *http.Request) {
	var (
		pname = r.URL.Query().Get("name")
		popt  = r.URL.Query().Get("opt")
		pdata = r.URL.Query().Get("data")
		pauth = r.URL.Query().Get("auth")
	)

	//check param
	if pauth != *auth || strings.ToUpper(popt) != "PUT" {
		w.WriteHeader(404)
		return
	}

	//push data to nsq
	if err = PushData(pname, []byte(pdata)); err != nil {
		log.Println(err)
	}
}

func PushData(topic string, body []byte) error {
	return nsqpro.Publish(topic, body)
}

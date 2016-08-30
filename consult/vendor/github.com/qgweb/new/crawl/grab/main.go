package main

import (
	"flag"
	"github.com/qgweb/new/crawl/grab/model"
)

var (
	nsqhost    = flag.String("nsq-host", "127.0.0.1", "nsq地址")
	nsqport    = flag.String("nsq-port", "4150", "nsq端口")
	receivekey = flag.String("rkey", "zhejiang_taglist_ad", "nsq接收的topic")
	sendkey    = flag.String("skey", "zhejiang_goodsqueue", "nsq发送的topic")
	gtype      = flag.String("gtype", "taobao", "抓取类型")
)

func init() {
	flag.Parse()
}

func main() {
	config := model.NsqConfig{}
	config.NsqHost = *nsqhost
	config.NsqPort = *nsqport
	config.ReceiveKey = *receivekey
	config.SendKey = *sendkey

	var cl model.Crawler
	var ds model.DataStreamer

	switch *gtype {
	case "taobao":
		cl = model.TaobaoCrawl{}
	case "jd":
		cl = model.JDCrawl{}
	}

	ds = model.NewNSQDataStream(config, cl)
	ds.Receive()
}

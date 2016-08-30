package model

import (
	"encoding/json"
	"fmt"
	"github.com/ngaut/log"
	"github.com/nsqio/go-nsq"
	"github.com/qgweb/new/lib/encrypt"
	"math/rand"
	"net/url"
	"strings"
	"time"
)

type NsqConfig struct {
	NsqHost    string
	NsqPort    string
	ReceiveKey string
	SendKey    string
}

type NSQDataStream struct {
	config      NsqConfig
	nsqproducer *nsq.Producer
	crawler     Crawler
}

func NewNSQDataStream(config NsqConfig, crawler Crawler) *NSQDataStream {
	ds := &NSQDataStream{
		config:  config,
		crawler: crawler,
	}
	ds.initNsqConn()
	return ds
}

func (this *NSQDataStream) initNsqConn() {
	var err error
	this.nsqproducer, err = nsq.NewProducer(fmt.Sprintf("%s:%s", this.config.NsqHost,
		this.config.NsqPort), nsq.NewConfig())
	if err != nil {
		log.Fatal("连接nsq出错,错误信息为:", err)
		return
	}
}
func (this *NSQDataStream) Receive() {
	cfg := nsq.NewConfig()
	cus, err := nsq.NewConsumer(this.config.ReceiveKey, this.config.ReceiveKey, cfg)
	if err != nil {
		log.Fatal("连接nsq失败")
	}

	cus.AddHandler(this)
	log.Warn(cus.ConnectToNSQD(fmt.Sprintf("%s:%s", this.config.NsqHost, this.config.NsqPort)))
	select {
	case <-cus.StopChan:
		return
	}
}

func (this *NSQDataStream) HandleMessage(message *nsq.Message) error {
	data := string(message.Body)
	if data == "" {
		log.Warn("数据丢失")
		return nil
	}

	data = encrypt.GetEnDecoder(encrypt.TYPE_BASE64).Decode(data)

	urlData, err := url.ParseQuery(data)
	if err != nil {
		log.Warn("解析数据失败")
		return nil
	}

	//数据存放在队列中
	if urlData.Get("date") != time.Now().Format("2006-01-02") {
		return nil
	}

	this.Dispatch(urlData)

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	_=time.Millisecond * time.Duration(r.Intn(1000) + 200)
	time.Sleep(time.Microsecond)
	return nil
}

func (this *NSQDataStream) Dispatch(data url.Values) {
	var (
		uidsAry  = strings.Split(data.Get("uids"), ",")
		info     = make(map[string]interface{})
		ginfoAry = this.GrabData(uidsAry)
	)

	data.Set("ua", encrypt.DefaultBase64.Decode(data.Get("ua")))

	for k, _ := range data {
		info[k] = data.Get(k)
	}

	if len(ginfoAry) == 0 {
		return
	}

	info["ginfos"] = ginfoAry

	j, err := json.Marshal(&info)
	if err != nil {
		log.Warn(err)
		return
	}

	go this.Save(j)
}

func (this *NSQDataStream) GrabData(gids []string) []map[string]interface{} {
	var (
		lenGids = len(gids)
		gChan   = make(chan map[string]interface{}, lenGids)
		cChan   = make(chan byte, 10)
		result  = make([]map[string]interface{}, 0, lenGids)
	)

	for i := 0; i < lenGids; i++ {
		go func(gid string) {
			defer func() {
				<-cChan
			}()
			cChan <- 1
			gChan <- this.crawler.Grab(gid)
		}(gids[i])
	}
	for i := 0; i < lenGids; i++ {
		v := <-gChan
		if v == nil {
			continue
		}

		result = append(result, v)
	}
	return result
}

func (this *NSQDataStream) Save(data []byte) {
	err := this.nsqproducer.Ping()
	if err != nil {
		log.Warn("无法和nsq通讯,错误信息为:", err)
		return
	}

	err = this.nsqproducer.Publish(this.config.SendKey, data)
	if err != nil {
		log.Warn("推送数据失败,错误信息为:", err)
	}

	//数据冗余处理（临时）
	err = this.nsqproducer.Publish(this.config.SendKey + "_es", data)
	if err != nil {
		log.Warn("推送数据失败,错误信息为:", err)
	}
}

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/url"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/ngaut/log"
	"github.com/nsqio/go-nsq"

	"github.com/qgweb/gopro/lib/encrypt"

	"gopkg.in/ini.v1"
	"gopkg.in/mgo.v2"
)

type QueueData struct {
	Data url.Values
	Px   string
}

type MapGoods struct {
	sync.Mutex
	Goods []map[string]interface{}
}

func NewMapGoods(length int) *MapGoods {
	return &MapGoods{Goods: make([]map[string]interface{}, 0, length)}
}

var (
	iniFile     *ini.File
	conf        = flag.String("conf", "conf.ini", "conf.ini")
	err         error
	queueSize   = 1
	dataQueue   = make(chan QueueData, queueSize)
	grabFactory *GrabFactory
)

func init() {
	flag.Parse()
	d, err := ioutil.ReadFile(*conf)
	if err != nil {
		log.Fatal("读取配置文件失败")
	}

	iniFile, err = ini.Load(d)
	if err != nil {
		log.Fatal("读取配置文件内容失败,错误信息为:", err)
	}

	initCateInfo()
	initHttpProxy()
	initNsqConn()
}

func bootstrap(px string) {
	for {
		data := httpsqsQueue(px)
		if data == nil {
			time.Sleep(time.Minute)
			continue
		}

		//数据存放在队列中
		nt := time.NewTicker(time.Minute * 10)
		select {
		case dataQueue <- QueueData{Data: data, Px: px}:
		case <-nt.C:
			log.Warn("队列超时")
		}

		seed := time.Duration(rand.New(rand.NewSource(time.Now().UnixNano())).Intn(200))
		time.Sleep(time.Millisecond * seed)
	}
}

// nsq版本
func bootstrapNsq(px string) {
	var (
		topic = iniFile.Section("queuekey").Key("key").String()
		host  = iniFile.Section("nsq").Key("host").String()
		port  = iniFile.Section("nsq").Key("port").String()
	)

	cus, err := nsq.NewConsumer(topic+px, topic, nsq.NewConfig())
	if err != nil {
		log.Fatal("连接nsq失败")
	}

	cus.AddHandler(NSQHandler{Px: px})
	log.Warn(cus.ConnectToNSQD(fmt.Sprintf("%s:%s", host, port)))
	select {
	case <-cus.StopChan:
		return
	}
}

// 分配数据
func dispath(data url.Values, px string) {
	if px == "_ad" && data.Get("date") != time.Now().Format("2006-01-02") {
		//return
	}

	data.Set("ua", encrypt.DefaultBase64.Decode(data.Get("ua")))
	uidsAry := strings.Split(data.Get("uids"), ",")

	ginfoAry := NewMapGoods(len(uidsAry))

	info := make(map[string]interface{})
	log.Error(len(uidsAry))
	bt := time.Now()
	wg := sync.WaitGroup{}

	for i := 0; i < len(uidsAry); i++ {
		wg.Add(1)
		go func(gid string) {
			defer func() {
				if msg := recover(); msg != nil {
					log.Error(msg)
				}
			}()

			//判断是否存在
			ginfo, err := checkGoodsExist(gid)
			if err == mgo.ErrNotFound {
				//抓取
				ginfo = GrabGoodsInfo(gid)
				if ginfo != nil {
					ginfo["exists"] = 0
				}
			} else {
				ginfo["exists"] = 1
			}

			ginfoAry.Lock()
			ginfoAry.Goods = append(ginfoAry.Goods, ginfo)
			ginfoAry.Unlock()
			wg.Done()
		}(uidsAry[i])
	}

	wg.Wait()
	log.Debug(time.Now().Sub(bt).Seconds())

	for k, _ := range data {
		info[k] = data.Get(k)
	}

	info["ginfos"] = ginfoAry.Goods

	j, err := json.Marshal(&info)
	if err != nil {
		log.Warn("err")
		return
	}

	pushMsgToNsq(j)
}

//循环读取队列
func loopQueue() {
	wg := sync.WaitGroup{}
	for {
		if len(dataQueue) == queueSize {
			for i := 0; i < queueSize; i++ {
				wg.Add(1)
				go func() {
					d := <-dataQueue
					dispath(d.Data, d.Px)
					wg.Done()
				}()
			}
			wg.Wait()
			r := rand.New(rand.NewSource(time.Now().UnixNano()))
			time.Sleep(time.Millisecond * time.Duration(r.Intn(1500)+1000))
		}
	}
}

func main() {
	log.SetHighlighting(true)
	runtime.GOMAXPROCS(runtime.NumCPU())
	//go bootstrap("_ad")
	//go bootstrap("_ck")
	go bootstrapNsq("_ad")
	go bootstrapNsq("_ck")
	go loopQueue()
	//go checkProxyHealth()
	select {}
}

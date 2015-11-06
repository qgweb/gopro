package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"math"
	"net/url"
	"runtime"
	"strings"
	"sync"

	"github.com/ngaut/log"
	"github.com/nsqio/go-nsq"

	"github.com/qgweb/gopro/lib/encrypt"

	"sync/atomic"

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
	grabFactory *GrabFactory
	recvCount   uint64
	dealCount   uint64
	isAliyun    bool //是否是阿里云机器
)

func initConfigFile() {
	flag.Parse()
	d, err := ioutil.ReadFile(*conf)
	if err != nil {
		log.Fatal("读取配置文件失败")
	}

	iniFile, err = ini.Load(d)
	if err != nil {
		log.Fatal("读取配置文件内容失败,错误信息为:", err)
	}

	isAliyun = iniFile.Section("default").Key("mode").String() == "1"
}

func init() {
	initConfigFile()
	initCateInfo()
	initNsqConn()
}

// nsq版本
func bootstrapNsq(px string) {
	var (
		topic = iniFile.Section("queuekey").Key("key").String()
		host  = iniFile.Section("nsq").Key("host").String()
		port  = iniFile.Section("nsq").Key("port").String()
	)

	cfg := nsq.NewConfig()

	cus, err := nsq.NewConsumer(topic+px, topic, cfg)
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

// 分页抓取商品
func grabPage(uidsAry []string) *MapGoods {
	var (
		uidsLen   = len(uidsAry)
		ginfoAry  = NewMapGoods(uidsLen)
		wg        = sync.WaitGroup{}
		uuidsAry  = uidsAry
		pageSize  = 20
		page      = 1
		pageCount = int(math.Ceil(float64(uidsLen) / float64(pageSize)))
	)

	if !isAliyun {
		uuidsAry = make([]string, 0, len(uidsAry))
		for i := 0; i < len(uidsAry); i++ {
			ginfo, err := checkGoodsExist(uidsAry[i])
			if err == mgo.ErrNotFound {
				uuidsAry = append(uuidsAry, uidsAry[i])
			} else {
				ginfo["exists"] = 1
				ginfoAry.Goods = append(ginfoAry.Goods, ginfo)
			}
		}

		pageCount = int(math.Ceil(float64(len(uuidsAry)) / float64(pageSize)))
	}

	for ; page <= pageCount; page++ {
		begin := (page - 1) * pageSize
		end := begin + pageSize
		if page == pageCount {
			end = uidsLen
		}

		for _, v := range uuidsAry[begin:end] {
			wg.Add(1)
			go func(gid string) {
				defer func() {
					wg.Done()
					if msg := recover(); msg != nil {
						log.Error(msg)
					}
				}()

				//抓取数据
				ginfo := GrabGoodsInfo(gid)
				if ginfo != nil {
					ginfo["exists"] = 0
				}

				if ginfo != nil {
					ginfoAry.Lock()
					ginfoAry.Goods = append(ginfoAry.Goods, ginfo)
					ginfoAry.Unlock()
				}
			}(v)
		}
		wg.Wait()
	}
	return ginfoAry
}

// 分配数据
func dispath(data url.Values, px string) {
	var (
		uidsAry  = strings.Split(data.Get("uids"), ",")
		info     = make(map[string]interface{})
		ginfoAry = grabPage(uidsAry)
	)

	data.Set("ua", encrypt.DefaultBase64.Decode(data.Get("ua")))

	for k, _ := range data {
		info[k] = data.Get(k)
	}

	if len(ginfoAry.Goods) == 0 {
		return
	}

	info["ginfos"] = ginfoAry.Goods

	j, err := json.Marshal(&info)
	if err != nil {
		log.Warn(err)
		return
	}

	pushMsgToNsq(j)
	dealCount = atomic.AddUint64(&dealCount, 1)
}

func main() {
	log.SetHighlighting(false)
	runtime.GOMAXPROCS(runtime.NumCPU())

	go bootstrapNsq("_ad")
	go bootstrapNsq("_ck")

	if !isAliyun {
		go MonitorLoop()
	}

	select {}
}

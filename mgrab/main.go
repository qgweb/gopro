package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/bitly/go-nsq"
	"goclass/convert"
	"io/ioutil"
	"log"
	"math/rand"
	"net/url"
	"runtime"
	"strings"
	"sync"
	"time"

	"goclass/encrypt"

	"gopkg.in/ini.v1"
	"gopkg.in/mgo.v2"
)

var (
	iniFile *ini.File
	conf    = flag.String("conf", "conf.ini", "conf.ini")
	err     error
)

func init() {
	flag.Parse()
	d, err := ioutil.ReadFile(*conf)
	if err != nil {
		log.Fatalln("读取配置文件失败")
	}

	iniFile, err = ini.Load(d)
	if err != nil {
		log.Fatalln("读取配置文件内容失败,错误信息为:", err)
	}

	initCateInfo()
	initUserAgent()
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

		dispath(data, px)
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
		log.Fatalln("连接nsq失败")
	}

	cus.AddHandler(NSQHandler{Px: px})
	cus.ConnectToNSQD(fmt.Sprintf("%s:%s", host, port))
	select {
	case <-cus.StopChan:
		return
	}
}

func dispath(data url.Values, px string) {
	if px == "_ad" && data.Get("date") != time.Now().Format("2006-01-02") {
		return
	}

	data.Set("ua", encrypt.GetEnDecoder(encrypt.TYPE_BASE64).Decode(data.Get("ua")))
	uidsAry := strings.Split(data.Get("uids"), ",")

	wg := WaitGroup{}
	ginfoAry := make([]map[string]interface{}, 0, len(uidsAry))
	info := make(map[string]interface{})
	var lock sync.Mutex

	for i := 0; i < len(uidsAry); i++ {
		wg.Wrap(func(param ...interface{}) {
			gid := param[0]
			//判断是否存在
			ginfo, err := checkGoodsExist(gid.(string))
			if err == mgo.ErrNotFound {
				//抓取
				ginfo = GrabGoodsInfo(gid.(string))
				if ginfo != nil {
					ginfo["exists"] = 0
				}
			} else {
				ginfo["exists"] = 1
			}
			if ginfo != nil {
				lock.Lock()
				ginfoAry = append(ginfoAry, ginfo)
				lock.Unlock()
			}
		}, uidsAry[i])
	}

	wg.Wait()

	for k, _ := range data {
		info[k] = data.Get(k)
	}

	info["ginfos"] = ginfoAry

	j, err := json.Marshal(&info)
	if err != nil {
		log.Println("err")
		return
	}

	pushMsgToNsq(j)
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	//go bootstrap("_ad")
	//go bootstrap("_ck")
	go bootstrapNsq("_ad")
	go bootstrapNsq("_ck")
	go checkProxyHealth()
	select {}
}

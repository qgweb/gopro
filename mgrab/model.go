package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/ngaut/log"

	"github.com/qgweb/gopro/lib/encrypt"
	"github.com/qgweb/gopro/lib/grab"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"sync"

	"net"
	"sync/atomic"

	"github.com/astaxie/beego/httplib"
	"github.com/nsqio/go-nsq"
	"github.com/qgweb/gopro/lib/convert"
)

var (
	mux         sync.Mutex
	mdbsession  *mgo.Session
	sexmap      map[string]int = map[string]int{"中性": 0, "男": 1, "女": 2}
	peoplemap   map[string]int = map[string]int{"青年": 0, "孕妇": 1, "中老年": 2, "儿童": 3, "青少年": 4, "婴儿": 5}
	cateList    map[string]map[string]interface{}
	nsqproducer *nsq.Producer
	useragents  []string
	httpproxys  []string
)

type WaitGroup struct {
	sync.WaitGroup
}

func (this *WaitGroup) Wrap(f func(p ...interface{}), param ...interface{}) {
	this.Add(1)
	go func() {
		f(param...)
		this.Done()
	}()
}

// 主机信息
type Hosts struct {
	Name   string   `json:"name"`
	Ip     string   `json:"ip"`
	Pid    string   `json:"pid"` //区分同一个ip多台机器
	Info   InfoData `json:"infodata"`
	Uptime int64    `json:"time"` // 更新时间
}

// 反馈的数据
type InfoData struct {
	Type       string `json:"type"` //类型
	ReceiveNum int    `json:"rnum"` //接收数据
	DealNum    int    `json:"dnum"` //处理数据
}

type NSQHandler struct {
	Px string
}

func (this NSQHandler) HandleMessage(message *nsq.Message) error {
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

	recvCount = atomic.AddUint64(&recvCount, 1)
	dispath(urlData, this.Px)
	return nil
}

//添加商品
func GrabGoodsInfo(gid string) (info map[string]interface{}) {
LABEL:
	url := "https://item.taobao.com/item.htm?id=" + gid
	h := grab.GrabTaoHTML(url)

	if h == "" {
		return nil
	}

	p, _ := grab.ParseNode(h)

	//标签名称
	title := grab.GetTitle(p)

	if title == "淘宝网 - 淘！我喜欢" || strings.Contains(title, "出错啦！") {
		//log.Println("商品不存在,id为:", gid)
		return nil
	}

	if strings.Contains(title, "访问受限") {
		log.Error("访问受限,id为", gid)
		time.Sleep(time.Minute * 2)
		goto LABEL
		return nil
	}

	//标签id
	cateId := grab.GetCategoryId(h)

	//标签信息

	cateInfo := make(map[string]interface{})

	if v, ok := cateList[cateId]; !ok {
		return nil
	} else {
		cateInfo = v
	}

	//特性
	features := make(map[string]int)
	if v, ok := cateInfo["features"]; ok {
		for a, b := range v.(map[string]interface{}) {
			features[a] = b.(int)
		}
	}

	//属性
	attrbuites := grab.GetAttrbuites(p)

	//性别
	sex := 0
	for k, v := range sexmap {
		if strings.Contains(title, k) {
			sex = v
			break
		}
	}
	//人群
	people := 0
	for k, v := range peoplemap {
		if strings.Contains(title, k) {
			people = v
			break
		}
	}

	// 店铺信息
	shopId := grab.GetShopId(p)
	shopName := grab.GetShopName(p)
	shopUrl := grab.GetShopUrl(p)
	shopBoss := grab.GetShopBoss(p)

	return map[string]interface{}{
		"shop_id":    shopId,
		"shop_name":  shopName,
		"shop_url":   shopUrl,
		"shop_boss":  shopBoss,
		"gid":        gid,
		"tagname":    cateInfo["name"],
		"tagid":      cateId,
		"features":   features,
		"attrbuites": attrbuites,
		"sex":        sex,
		"people":     people,
	}
}

//获取mongo数据库链接
func GetSession() *mgo.Session {
	mux.Lock()
	defer mux.Unlock()

	var (
		mouser = iniFile.Section("mongo").Key("user").String()
		mopwd  = iniFile.Section("mongo").Key("pwd").String()
		mohost = iniFile.Section("mongo").Key("host").String()
		moport = iniFile.Section("mongo").Key("port").String()
		modb   = iniFile.Section("mongo").Key("db").String()
	)

	if mdbsession == nil {
		var err error
		mdbsession, err = mgo.Dial(fmt.Sprintf("%s:%s@%s:%s/%s", mouser, mopwd, mohost, moport, modb))
		if err != nil {
			panic(err)
		}
	}
	//高并发下会关闭连接,ping下会恢复
	mdbsession.Ping()
	return mdbsession.Copy()
}

// 初始化分类信息
func initCateInfo() {
	sess := GetSession()
	defer sess.Close()

	var (
		modb = iniFile.Section("mongo").Key("db").String()
	)

	var list []map[string]interface{}
	cateList = make(map[string]map[string]interface{})

	err := sess.DB(modb).C("taocat").Find(bson.M{"type": "0"}).Select(bson.M{"_id": 0}).All(&list)
	if err == mgo.ErrNotFound {
		err = sess.DB(modb).C("taocat").Find(bson.M{}).Select(bson.M{"_id": 0}).All(&list)
		if err == mgo.ErrNotFound {
			log.Error("读取淘宝分类出错")
		}
	}

	for _, v := range list {
		cateList[v["cid"].(string)] = v
	}
}

func initNsqConn() {
	var (
		host = iniFile.Section("nsq").Key("host").String()
		port = iniFile.Section("nsq").Key("port").String()
		err  error
	)
	nsqproducer, err = nsq.NewProducer(fmt.Sprintf("%s:%s", host, port), nsq.NewConfig())
	if err != nil {
		log.Fatal("连接nsq出错,错误信息为:", err)
		return
	}
}

//推送数据
func pushMsgToNsq(data []byte) {
	var (
		key = iniFile.Section("nsq").Key("key").String()
	)

	err := nsqproducer.Ping()
	if err != nil {
		log.Warn("无法和nsq通讯,错误信息为:", err)
		return
	}

	err = nsqproducer.Publish(key, data)
	if err != nil {
		log.Warn("推送数据失败,错误信息为:", err)
	}
}

//判断是否存在该商品
func checkGoodsExist(gid string) (res map[string]interface{}, err error) {
	sess := GetSession()
	defer func() {
		sess.Close()
	}()

	var (
		modb = iniFile.Section("mongo").Key("db").String()
	)

	info := make(map[string]interface{})

	err = sess.DB(modb).C("goods").Find(bson.M{"gid": gid}).
		Select(bson.M{"_id": 0}).One(&info)

	if err == mgo.ErrNotFound {
		return nil, err
	}

	//更新count字段
	sess.DB(modb).C("goods").Update(bson.M{"gid": gid}, bson.M{"$inc": bson.M{"count": 1}})

	if _, ok := info["shop_id"]; !ok {
		return nil, mgo.ErrNotFound
	}

	info1 := make(map[string]string)

	//获取店铺信息
	sess.DB(modb).C("taoshop").Find(bson.M{"shop_id": info["shop_id"]}).
		Select(bson.M{"_id": 0}).One(&info1)

	if _, ok := info1["shop_id"]; !ok {
		return nil, mgo.ErrNotFound
	}

	for k, v := range info1 {
		info[k] = v
	}

	return info, nil
}

func initHttpProxy() {
	httpproxys = make([]string, 0, 10)
	keys := iniFile.Section("httpproxy").Keys()
	for _, v := range keys {
		httpproxys = append(httpproxys, "http://"+v.String())
	}
}

func getHttpProxy() string {
	if len(httpproxys) == 0 {
		return ""
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	index := r.Intn(len(httpproxys))
	return httpproxys[index]
}

func checkProxyHealth() {
	for {
		select {
		case <-time.After(time.Second * 2):
			tmpProxy := make([]string, 0, 10)
			for _, v := range httpproxys {
				res, err := http.Get(v)

				if err == nil && res != nil {
					if res.StatusCode == 500 {
						a, err := ioutil.ReadAll(res.Body)
						if err == nil {
							if strings.Contains(string(a), "This is a proxy server") {
								tmpProxy = append(tmpProxy, v)
							}
						}
					}
				}
			}
			httpproxys = tmpProxy
		}
	}
}

func initGrabFactory() {
	grabFactory = NewFactory(200, 100)
}

func getIp() string {
	if conn, err := net.Dial("udp", "www.baidu.com:80"); err == nil {
		ip := conn.LocalAddr().String()
		conn.Close()
		return strings.Split(ip, ":")[0]
	}
	return ""
}

func MonitorLoop() {
	t := time.NewTicker(time.Second * 10)
	for {
		select {
		case <-t.C:
			var (
				host = iniFile.Section("monitor").Key("host").String()
				port = iniFile.Section("monitor").Key("port").String()
				url  = fmt.Sprintf("http://%s:%s/add", host, port)
			)
			data := Hosts{}
			data.Name, _ = os.Hostname()
			data.Ip = getIp()
			data.Pid = convert.ToString(os.Getpid())
			data.Info.DealNum = int(atomic.LoadUint64(&dealCount))
			data.Info.ReceiveNum = int(atomic.LoadUint64(&recvCount))
			data.Info.Type = "mgrab"
			by, err := json.Marshal(&data)
			if err == nil {
				http.Post(url, "application/json", ioutil.NopCloser(bytes.NewReader(by)))
			}

		}
	}
}

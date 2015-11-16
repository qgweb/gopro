package main

import (
	"encoding/json"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/ngaut/log"

	gs "github.com/qgweb/gopro/lib/grab"

	"bytes"
	"io/ioutil"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/qgweb/gopro/lib/convert"
	"github.com/qgweb/gopro/lib/encrypt"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var (
	mdbsession    *mgo.Session
	mdbputsession *mgo.Session
	mux           sync.Mutex
)

type Goods struct {
	Gid        string         `json:"gid"`
	Tagname    string         `json:"tagname"`
	Tagid      string         `json:"tagid"`
	Features   map[string]int `json:"features"`
	Attrbuites string         `json:"attrbuites"`
	Sex        int            `json:"sex"`
	People     int            `json:"people"`
	Shop_id    string         `json:"shop_id"`
	Shop_name  string         `json:"shop_name"`
	Shop_url   string         `json:"shop_url"`
	Shop_boss  string         `json:"shop_boss"`
	Count      int            `json:"count"`
	Exists     int            `json:"exists"`
}

type CombinationData struct {
	Ad     string  `json:"ad"`
	Cookie string  `json:"cookie"`
	Ua     string  `json:"ua"`
	Clock  string  `json:"clock"`
	Date   string  `json:"date"`
	Uids   string  `json:"uids"`
	Ginfos []Goods `json:"ginfos"`
}

type UserLocus struct {
	AD     string
	UA     string
	Hour   string
	TagIds []string
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

//分配数据
func dispath(data *CombinationData) {
	defer func() {
		if msg := recover(); msg != nil {
			fmt.Println(msg)
		}
	}()

	var (
		goodsLen = len(data.Ginfos)
		cids     = make([]string, 0, goodsLen)
		shopids  = make([]string, 0, goodsLen)
		cf       = NewChanFactory(5, goodsLen)
	)

	for _, g := range data.Ginfos {
		cids = append(cids, g.Tagid)
		shopids = append(shopids, g.Shop_id)
		f := ChanFunction{}
		f.Fun = func(p ...interface{}) {
			gs := p[0].(Goods)
			if gs.Exists == 0 {
				addGoods(gs)
			}
			addShop(gs)
		}
		f.Params = []interface{}{g}
		cf.Push(f)
	}

	cf.Run()

	AddUidCids(map[string]string{
		"ad": data.Ad, "cookie": data.Cookie,
		"ua": data.Ua, "cids": strings.Join(cids, ","),
		"shops": strings.Join(shopids, ","),
		"clock": data.Clock, "date": data.Date})
	dealCount = atomic.AddUint64(&dealCount, 1)
}

// 添加店铺
func addShop(g Goods) {
	sess := GetSession()

	var (
		modb = iniFile.Section("mongo").Key("db").String()
	)

	sess.DB(modb).C("taoshop").Upsert(bson.M{"shop_id": g.Shop_id}, bson.M{
		"$set": bson.M{
			"shop_name": g.Shop_name,
			"shop_url":  g.Shop_url,
			"shop_boss": g.Shop_boss,
		},
	})

	sess.Close()
}

//添加商品
func addGoods(g Goods) {
	var (
		modb = iniFile.Section("mongo").Key("db").String()
		sess = GetSession()
	)

	sess.DB(modb).C("goods").Upsert(bson.M{"gid": g.Gid}, bson.M{"$set": bson.M{
		"tagname": g.Tagname, "tagid": g.Tagid, "features": g.Features,
		"attrbuites": g.Attrbuites, "sex": g.Sex, "people": g.People,
		"shop_id": g.Shop_id, "shop_name": g.Shop_name, "shop_url": g.Shop_url,
		"shop_box": g.Shop_boss, "count": g.Count}})

	sess.Close()
}

//添加用户id对应分类id
func AddUidCids(param map[string]string) {
	//ad string, cids string, cookie string, ua string
	var (
		modb   = iniFile.Section("mongo").Key("db").String()
		prefix = iniFile.Section("queuekey").Key("prefix").String()
	)

	sess := GetSession()
	defer func() {
		sess.Close()
	}()

	tableName := prefix + "ad_tags"
	if param["cookie"] != "" {
		tableName = prefix + "cookie_tags"
	}

	//无cookie情况
	if param["cookie"] == "" {
		sess.DB(modb).C(tableName).Upsert(bson.M{"ad": param["ad"], "ua": param["ua"]},
			bson.M{"$set": bson.M{"cids": param["cids"]}})

		//按小时存储
		t := tableName + "_clock"
		c := "cids." + param["clock"]
		d := param["date"]
		sess.DB(modb).C(t).Upsert(bson.M{"ad": param["ad"], "ua": param["ua"], "date": d},
			bson.M{"$set": bson.M{c: param["cids"]}})

		//合并到用户轨迹上
		userLocusCombine(UserLocus{
			AD:     param["ad"],
			Hour:   param["clock"],
			TagIds: strings.Split(param["cids"], ","),
			UA:     encrypt.DefaultBase64.Encode(param["ua"]),
		})

		//存储对应的店铺信息
		t = tableName + "_shop"
		for _, v := range strings.Split(param["shops"], ",") {
			sess.DB(modb).C(t).Upsert(bson.M{"ad": param["ad"], "ua": param["ua"], "date": d},
				bson.M{"$addToSet": bson.M{"shop": bson.M{"id": v}}})
		}
	} else {
		sess.DB(modb).C(tableName).Upsert(bson.M{"cookie": param["cookie"]},
			bson.M{"$set": bson.M{"cids": param["cids"], "ad": param["ad"]}})

		//统计标签频率
		tagFrequencyRecord(param["cookie"], param["cids"])
	}
}

// 标签频率统计记录
func tagFrequencyRecord(cookie string, cids string) {
	var (
		modb   = iniFile.Section("mongo").Key("db").String()
		prefix = iniFile.Section("queuekey").Key("prefix").String()
	)

	sess := GetSession()
	defer sess.Close()

	//分割标签
	tagAry := strings.Split(cids, ",")
	tagsMap := make(map[string]int)

	for _, v := range tagAry {
		if _, ok := tagsMap[v]; ok {
			tagsMap[v] = tagsMap[v] + 1
		} else {
			tagsMap[v] = 1
		}
	}

	//排序
	s := gs.NewMapSorter(tagsMap)
	s.Sort()

	bms := make([]bson.M, 0, 20)
	for _, v := range s {
		bms = append(bms, bson.M{"tagid": v.Key, "tagcount": v.Val})
	}

	//插入mongo
	tableName := prefix + "cookie_tags_put"
	sess.DB(modb).C(tableName).Upsert(bson.M{"cookie": cookie},
		bson.M{"cookie": cookie, "cids": bms, "date": time.Now().Format("2006-01-02 15:04:05")})
}

// 用户电商轨迹合并到投放用户轨迹
func userLocusCombine(ul UserLocus) {
	sess := GetPutSession()
	defer sess.Close()

	var (
		modb    = iniFile.Section("mongo-put").Key("db").String()
		prefix  = iniFile.Section("queuekey").Key("prefix").String()
		motable = "useraction"
		tags    = make([]bson.M, 0, len(ul.TagIds))
		day     = time.Now().Format("20060102")
		info    map[string]interface{}
	)

	if prefix == "jiangsu_" {
		motable = motable + "_jiangsu"
	}

	for _, v := range ul.TagIds {
		tags = append(tags, bson.M{
			"tagmongo": "0",
			"tagId":    v,
			"tagType":  "1",
		})
	}

	err := sess.DB(modb).C(motable).Find(bson.M{"AD": ul.AD, "UA": ul.UA,
		"hour": ul.Hour, "day": day}).One(&info)
	if err == mgo.ErrNotFound {
		//插入
		sess.DB(modb).C(motable).Insert(bson.M{
			"domainId": "0",
			"AD":       ul.AD,
			"UA":       ul.UA,
			"hour":     ul.Hour,
			"day":      time.Now().Format("20060102"),
			"tag":      tags,
		})
	}

	if err == nil {
		//更新
		sess.DB(modb).C(motable).Upsert(bson.M{"AD": ul.AD, "UA": ul.UA, "hour": ul.Hour,
			"day": day}, bson.M{"$push": bson.M{"tag": bson.M{"$each": tags}}})
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
			log.Fatal(err)
		}
	}
	//高并发下会关闭连接,ping下会恢复
	mdbsession.Ping()
	return mdbsession.Copy()
}

//获取mongo数据库链接
func GetPutSession() *mgo.Session {
	mux.Lock()
	defer mux.Unlock()

	var (
		mouser = iniFile.Section("mongo-put").Key("user").String()
		mopwd  = iniFile.Section("mongo-put").Key("pwd").String()
		mohost = iniFile.Section("mongo-put").Key("host").String()
		moport = iniFile.Section("mongo-put").Key("port").String()
		modb   = iniFile.Section("mongo-put").Key("db").String()
	)

	if mdbputsession == nil {
		var err error

		mdbputsession, err = mgo.Dial(fmt.Sprintf("%s%s:%s/%s", func() string {
			if mouser == "" && mopwd == "" {
				return ""
			} else {
				return mouser + ":" + mopwd + "@"
			}
		}(), mohost, moport, modb))

		if err != nil {
			log.Fatal(err)
		}
	}
	//高并发下会关闭连接,ping下会恢复
	mdbputsession.Ping()
	return mdbputsession.Copy()
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
			data.Info.Type = "mtag"
			by, err := json.Marshal(&data)
			if err == nil {
				http.Post(url, "application/json", ioutil.NopCloser(bytes.NewReader(by)))
			}

		}
	}
}

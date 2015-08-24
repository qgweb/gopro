// taobao category service
package main

import (
	"fmt"
	"gopro/lib/grab"
	"log"
	"strings"
	"sync"

	"github.com/garyburd/redigo/redis"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var (
	mdbsession *mgo.Session
	mux        sync.Mutex
	sexmap     map[string]int = map[string]int{"中性": 0, "男": 1, "女": 2}
	peoplemap  map[string]int = map[string]int{"青年": 0, "孕妇": 1, "中老年": 2, "儿童": 3, "青少年": 4, "婴儿": 5}
)

//获取session
func GetSession() *mgo.Session {
	mux.Lock()
	defer mux.Unlock()

	if mdbsession == nil {
		var (
			err    error
			mouser = IniFile.Section("mongo-xu_precise").Key("user").String()
			mopwd  = IniFile.Section("mongo-xu_precise").Key("pwd").String()
			mohost = IniFile.Section("mongo-xu_precise").Key("host").String()
			moport = IniFile.Section("mongo-xu_precise").Key("port").String()
			modb   = IniFile.Section("mongo-xu_precise").Key("db").String()
		)

		mdbsession, err = mgo.Dial(fmt.Sprintf("%s:%s@%s:%s/%s", mouser, mopwd, mohost, moport, modb))
		if err != nil {
			panic(err)
		}
	}
	//高并发下会关闭连接,ping下会恢复
	mdbsession.Ping()
	return mdbsession.Copy()
}

//获取reids连接池
func RedisPool() *redis.Pool {
	var (
		host = IniFile.Section("redis").Key("host").String()
		port = IniFile.Section("redis").Key("port").String()
	)

	return &redis.Pool{
		MaxIdle:   80,
		MaxActive: 1200, // max number of connections
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", host+":"+port)
			if err != nil {
				panic(err.Error())
			}
			return c, err
		},
	}
}

//获取淘宝分类
func GetTaoCat(cid string) (map[string]interface{}, error) {
	modb := IniFile.Section("mongo-xu_precise").Key("db").String()
	sess := GetSession()
	defer sess.Close()
	info := make(map[string]interface{})
	err := sess.DB(modb).C("taocat").Find(bson.M{"cid": cid}).One(&info)
	if err != nil {
		return nil, err
	}
	return info, nil
}

//添加商品
func AddGoodsInfo(gid string) (cid string) {
	modb := IniFile.Section("mongo-xu_precise").Key("db").String()
	url := "http://item.taobao.com/item.htm?id=" + gid
	h := grab.GrabTaoHTML(url)
	p, _ := grab.ParseNode(h)
	sess := GetSession()

	//标签名称
	title := grab.GetTitle(p)
	if title == "淘宝网 - 淘！我喜欢" {
		return ""
	}

	//标签id
	cateId := grab.GetCategoryId(h)
	cateInfo, err := GetTaoCat(cateId)
	if err == mgo.ErrNotFound {
		log.Println("分类ID:", cateId, err)
		return ""
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
	//浏览数
	count := 0

	// 店铺信息
	shopId := grab.GetShopId(p)
	shopName := grab.GetShopName(p)
	shopUrl := grab.GetShopUrl(p)
	shopBoss := grab.GetShopBoss(p)

	go func() {
		sess.DB(modb).C("goods").Upsert(bson.M{"gid": gid}, bson.M{"gid": gid,
			"tagname": cateInfo["name"], "tagid": cateId, "features": features,
			"attrbuites": attrbuites, "sex": sex, "people": people, "shop_id": shopId,
			"shop_name": shopName, "shop_url": shopUrl, "shop_boss": shopBoss,
			"count": count,
		})
		sess.Close()
	}()

	return cateId
}

// taobao tag
type Taotag struct {
}

//获取标签
func (this Taotag) GetTag(gid string, taokeUrl string) []byte {
	modb := IniFile.Section("mongo-xu_precise").Key("db").String()
	sess := GetSession()
	defer sess.Close()
	info := make(map[string]interface{})
	err := sess.DB(modb).C("goods").Find(bson.M{"gid": gid}).
		Select(bson.M{"tagid": 1, "tagname": 1, "_id": 0}).One(&info)

	if err == mgo.ErrNotFound {
		cid := ""
		//淘宝客
		if taokeUrl != "" {
			h := grab.GrabTaoHTML(taokeUrl)
			cid, gid = grab.GetTaoCategoryId(h)
		} else {
			//抓取商品
			url := "http://item.taobao.com/item.htm?id=" + gid
			h := grab.GrabTaoHTML(url)
			cid = grab.GetCategoryId(h)
		}

		info, err := GetTaoCat(cid)
		if err != nil {
			return jsonReturn(`{"tagid" : "0", "tagname" : ""}`, err)
		}

		AddGoodsInfo(gid)

		return jsonReturn(map[string]interface{}{"tagid": cid, "tagname": info["name"]}, nil)

	} else if err == nil {
		return jsonReturn(info, nil)
	}

	return jsonReturn("", err)
}

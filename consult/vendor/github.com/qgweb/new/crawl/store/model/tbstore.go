package model

import (
	"encoding/json"
	"fmt"
	"github.com/ngaut/log"
	"github.com/qgweb/new/lib/mongodb"
	"github.com/qgweb/new/lib/timestamp"
	"time"
)

type Goods struct {
	Gid        string         `json:"gid",bson:"gid"`
	Tagname    string         `json:"tagname",bson:"tagname"`
	Tagid      string         `json:"tagid",bson:"tagid"`
	Features   map[string]int `json:"features",bson:"features"`
	Attrbuites string         `json:"attrbuites",bson:"attrbuites"`
	Sex        int            `json:"sex",json:"sex"`
	People     int            `json:"people",bson:"people"`
	Shop_id    string         `json:"shop_id",bson:"shop_id"`
	Shop_name  string         `json:"shop_name",bson:"shop_name"`
	Shop_url   string         `json:"shop_url",bson:"shop_url"`
	Shop_boss  string         `json:"shop_boss",bson:"shop_boss"`
	Brand      string         `json:"brand",bson:"brand"`
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

type TaobaoMongoStore struct {
	store  *mongodb.Mongodb
	put    *mongodb.Mongodb
	catMap map[string]string
	prefix string
}

func NewTaobaoMongoStore(c Config) *TaobaoMongoStore {
	var err error
	var m = mongodb.MongodbConf{}
	var tms = &TaobaoMongoStore{}
	m.Db = "xu_precise"
	m.Host = c.MgoStoreHost
	m.Port = c.MgoStorePort
	m.UName = c.mgoStoreUname
	m.Upwd = c.mgoStoreUpwd
	tms.store, err = mongodb.NewMongodb(m)
	if err != nil {
		log.Fatal(err)
	}

	m.Db = "data_source"
	m.Host = c.MgoPutHost
	m.Port = c.MgoPutPort
	m.UName = ""
	m.Upwd = ""
	tms.put, err = mongodb.NewMongodb(m)
	if err != nil {
		log.Fatal(err)
	}

	tms.prefix = c.TablePrefixe
	tms.initCategory()
	return tms
}

func (this *TaobaoMongoStore) initCategory() {
	q := mongodb.MongodbQueryConf{}
	q.Db = "xu_precise"
	q.Table = "taocat"
	q.Select = mongodb.MM{"name": 1, "cid": 1, "_id": 0}
	q.Query = mongodb.MM{"type": "0"}
	this.catMap = make(map[string]string)
	this.store.Query(q, func(info map[string]interface{}) {
		this.catMap[info["cid"].(string)] = info["name"].(string)
	})
}

func (this *TaobaoMongoStore) saveGoods(gs []Goods) {
	q := mongodb.MongodbQueryConf{}
	q.Db = "xu_precise"
	q.Table = "goods"

	for _, g := range gs {
		tagName := ""
		if v, ok := this.catMap[g.Tagid]; ok {
			tagName = v
		}
		q.Query = mongodb.MM{"gid": g.Gid}
		q.Update = mongodb.MM{"$set": mongodb.MM{
			"tagname": tagName, "tagid": g.Tagid, "features": g.Features,
			"attrbuites": g.Attrbuites, "sex": g.Sex, "people": g.People,
			"shop_id": g.Shop_id, "shop_name": g.Shop_name, "shop_url": g.Shop_url,
			"shop_box": g.Shop_boss}}
		this.store.Upsert(q)
	}
}

func (this *TaobaoMongoStore) saveAdTrace(cd *CombinationData) {
	t := this.prefix + "_ad_tags_clock"
	c := "cids." + cd.Clock
	cids := make([]string, 0, len(cd.Ginfos))
	for _, v := range cd.Ginfos {
		cids = append(cids, v.Tagid)
	}
	q := mongodb.MongodbQueryConf{}
	q.Db = "xu_precise"
	q.Table = t
	q.Query = mongodb.MM{"ad": cd.Ad, "ua": cd.Ua, "date": cd.Date}
	q.Update = mongodb.MM{"$set": mongodb.MM{c: cids}}
	this.store.Upsert(q)
}

func (this *TaobaoMongoStore) saveShopTrace(cd *CombinationData) {
	t := this.prefix + "_ad_shop_clock"
	q := mongodb.MongodbQueryConf{}
	q.Db = "xu_precise"
	q.Table = t
	q.Query = mongodb.MM{"ad": cd.Ad, "ua": cd.Ua, "date": cd.Date}
	for _, v := range cd.Ginfos {
		q.Update = mongodb.MM{"$addToSet": mongodb.MM{"shop": mongodb.MM{"id": v.Shop_id}}}
		this.store.Upsert(q)
	}
}

func (this *TaobaoMongoStore) pushTrace(cd *CombinationData) {
	t := "useraction"
	ts := timestamp.GetTimestamp(fmt.Sprintf("%s %s:%s:%s", cd.Date, cd.Clock, "00", "00"))
	tags := make([]mongodb.MM, 0, len(cd.Ginfos))

	if this.prefix == "jiangsu" {
		t = t + "_jiangsu"
	}

	for _, v := range cd.Ginfos {
		tags = append(tags, mongodb.MM{
			"tagmongo": "0",
			"tagId":    v.Tagid,
			"tagType":  "1",
		})
	}

	q := mongodb.MongodbQueryConf{}
	q.Db = "data_source"
	q.Table = t
	q.Query = mongodb.MM{"AD": cd.Ad, "UA": cd.Ua, "hour": cd.Clock, "day": time.Now().Format("20060102")}

	v, _ := this.put.Count(q)
	if v == 0 {
		q.Insert = []interface{}{
			mongodb.MM{
				"domainId":  "0",
				"AD":        cd.Ad,
				"UA":        cd.Ua,
				"tag":       tags,
				"hour":      cd.Clock,
				"day":       time.Now().Format("20060102"),
				"timestamp": ts,
			},
		}
		this.put.Insert(q)
	} else {
		q.Update = mongodb.MM{"$push": mongodb.MM{"tag": mongodb.MM{"$each": tags}}}
		this.put.Upsert(q)
	}
}

func (this *TaobaoMongoStore) ParseData(data interface{}) interface{} {
	cdata := &CombinationData{}
	err := json.Unmarshal(data.([]byte), cdata)
	if err != nil {
		log.Error("数据解析出错,错误信息为:", err)
		return nil
	}

	if cdata.Date != time.Now().Format("2006-01-02") {
		return nil
	}
	return cdata
}

func (this *TaobaoMongoStore) Save(info interface{}) {
	cd, ok := info.(*CombinationData)
	if !ok {
		return
	}
	this.saveGoods(cd.Ginfos)
	this.saveAdTrace(cd)
	this.saveShopTrace(cd)
	this.pushTrace(cd)
}

type TaoBaoDataStore struct {
	DataStore
}

func NewTaoBaoDataStore(c Config) *TaoBaoDataStore {
	tbs := &TaoBaoDataStore{}
	tbs.DataStore.sr = NewTaobaoMongoStore(c)
	return tbs
}

func NewTaoBaoEsDataStore(c Config) *TaoBaoDataStore {
	tbs := &TaoBaoDataStore{}
	tbs.DataStore.sr = NewTaobaoESStore(c)
	return tbs
}

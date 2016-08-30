package model

import (
	"encoding/json"
	"fmt"
	"github.com/gobuild/log"
	"github.com/qgweb/gopro/lib/encrypt"
	"github.com/qgweb/new/lib/mongodb"
	"github.com/qgweb/new/lib/timestamp"
	"gopkg.in/olivere/elastic.v3"
	"strings"
	"time"
)

type TaobaoESStore struct {
	client  *elastic.Client
	bulk    *elastic.BulkService
	store   *mongodb.Mongodb
	catMap  map[string]string
	prefix  string
	geohost string
}

func NewTaobaoESStore(c Config) *TaobaoESStore {
	var esstor = &TaobaoESStore{}
	var err error
	var m = mongodb.MongodbConf{}

	esstor.client, err = elastic.NewClient(elastic.SetURL(strings.Split(c.ESHost, ",")...))
	if err != nil {
		log.Fatal(err)
	}

	m.Db = "xu_precise"
	m.Host = c.MgoStoreHost
	m.Port = c.MgoStorePort
	m.UName = c.mgoStoreUname
	m.Upwd = c.mgoStoreUpwd
	esstor.store, err = mongodb.NewMongodb(m)
	if err != nil {
		log.Fatal(err)
	}
	esstor.prefix = c.TablePrefixe
	esstor.initCategory()
	esstor.geohost = c.GeoHost
	esstor.bulk = esstor.client.Bulk()
	return esstor
}

func (this *TaobaoESStore) initCategory() {
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
func (this *TaobaoESStore) saveGoods(gs []Goods) {
	var db = "taobao_goods"
	var table = "goods"
	for _, g := range gs {
		if v, ok := this.catMap[g.Tagid]; ok {
			g.Tagname = v
		}
		//this.client.Index().Index(db).Type(table).Id(g.Gid).BodyJson(g).Do()
		this.bulk.Add(elastic.NewBulkIndexRequest().Index(db).Type(table).Id(g.Gid).Doc(g))
	}
}

func (this *TaobaoESStore) saveAdTrace(cd *CombinationData) {
	var date = timestamp.GetTimestamp(fmt.Sprintf("%s %s:%s:%s", cd.Date, cd.Clock, "00", "00"))
	var id = encrypt.DefaultMd5.Encode(date + cd.Ad + cd.Ua)
	var db = this.prefix + "_tb_ad_trace"
	var table = "ad"
	var cids = make([]string, 0, len(cd.Ginfos))

	for _, v := range cd.Ginfos {
		cids = append(cids, v.Tagid)
	}

	//log.Info(this.client.Index().Index(db).Type(table).Id(id).BodyJson(map[string]interface{}{
	//	"ad":        cd.Ad,
	//	"ua":        cd.Ua,
	//	"timestamp": date,
	//	"cids":      cids,
	//}).Do())
	this.bulk.Add(elastic.NewBulkIndexRequest().Index(db).Type(table).Id(id).Doc(map[string]interface{}{
		"ad":        cd.Ad,
		"ua":        cd.Ua,
		"timestamp": date,
		"cids":      cids,
	}))
}

func (this *TaobaoESStore) saveShopTrace(cd *CombinationData) {
	var date = timestamp.GetTimestamp(fmt.Sprintf("%s %s:%s:%s", cd.Date, cd.Clock, "00", "00"))
	var db = this.prefix + "_tb_shop_trace"
	var id = encrypt.DefaultMd5.Encode(date + cd.Ad + cd.Ua)
	var table = "shop"
	var shopids = make([]string, 0, len(cd.Ginfos))
	for _, v := range cd.Ginfos {
		shopids = append(shopids, v.Shop_id)
	}

	//查询是否存在
	res, err := this.client.Search().Index(db).Type(table).Query(elastic.NewIdsQuery(table).Ids(id)).Fields("shop").Do()
	if err != nil {
		log.Error(err)
	}

	if res == nil || res.TotalHits() == 0 {
		//log.Info(this.client.Index().Index(db).Type(table).Id(id).BodyJson(map[string]interface{}{
		//	"ad":        cd.Ad,
		//	"ua":        cd.Ua,
		//	"timestamp": date,
		//	"shop":      shopids,
		//}).Do())
		this.bulk.Add(elastic.NewBulkIndexRequest().Index(db).Type(table).Id(id).Doc(
			map[string]interface{}{
				"ad":        cd.Ad,
				"ua":        cd.Ua,
				"timestamp": date,
				"shop":      shopids,
			}))
	} else {
		oshopids := res.Hits.Hits[0].Fields["shop"].([]interface{})
		var tmpMap = make(map[string]byte)
		for _, vv := range oshopids {
			tmpMap[vv.(string)] = 1
		}
		for _, vv := range shopids {
			tmpMap[vv] = 1
		}

		nshopids := make([]string, 0, len(tmpMap))
		for k, _ := range tmpMap {
			nshopids = append(nshopids, k)
		}

		//log.Info(this.client.Update().Index(db).Type(table).Doc(map[string]interface{}{
		//	"shop": nshopids,
		//}).Id(id).Do())
		this.bulk.Add(elastic.NewBulkUpdateRequest().Index(db).Type(table).Doc(map[string]interface{}{
			"shop": nshopids,
		}).Id(id))
	}
}

func (this *TaobaoESStore) getTagNames(goods []Goods) []string {
	var tns = make([]string, 0, len(goods))
	for _, g := range goods {
		if v, ok := this.catMap[g.Tagid]; ok {
			if strings.TrimSpace(v) != "" {
				tns = append(tns, v)
			}
		}
	}
	return tns
}

func (this *TaobaoESStore) getBrands(goods []Goods) []string {
	var tns = make([]string, 0, len(goods))
	for _, g := range goods {
		if strings.TrimSpace(g.Brand) != "" {
			tns = append(tns, g.Brand)
		}
	}
	return tns
}

func (this *TaobaoESStore) pushTagToMap(cd *CombinationData) {
	var (
		db1 = "map_trace"
		db2 = "map_trace_search"
		table = "map"
		date = timestamp.GetTimestamp(fmt.Sprintf("%s %s:%s:%s", cd.Date, cd.Clock, "00", "00"))
		id = encrypt.DefaultMd5.Encode(date + cd.Ad + encrypt.DefaultBase64.Encode(cd.Ua))
		tagNames = this.getTagNames(cd.Ginfos)
		brands = this.getBrands(cd.Ginfos)
		geo = GetLonLat(cd.Ad, this.geohost)
	)

	info := map[string]interface{}{
		"ad":        cd.Ad,
		"ua":        encrypt.DefaultBase64.Encode(cd.Ua),
		"timestamp": date,
		"tb_tags":   tagNames,
		"tb_brand":  brands,
		"geo":       geo,
	}

	if geo != "" {
		//this.client.Update().Index(db1).Type(table).Id(id).Doc(info).DocAsUpsert(true).Do()
		//this.client.Update().Index(db2).Type(table).Id(id).Doc(info).DocAsUpsert(true).Do()
		this.bulk.Add(elastic.NewBulkUpdateRequest().Index(db1).Type(table).Id(id).Doc(info).DocAsUpsert(true))
		this.bulk.Add(elastic.NewBulkUpdateRequest().Index(db2).Type(table).Id(id).Doc(info).DocAsUpsert(true))
	}
}

func (this *TaobaoESStore) ParseData(data interface{}) interface{} {
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

func (this *TaobaoESStore) Save(info interface{}) {
	cd, ok := info.(*CombinationData)
	if !ok {
		return
	}
	this.saveGoods(cd.Ginfos)
	this.saveAdTrace(cd)
	this.saveShopTrace(cd)
	//this.pushTagToMap(cd)
	if this.bulk.NumberOfActions() % 100 == 0 {
		log.Info(this.bulk.Do())
	}
}

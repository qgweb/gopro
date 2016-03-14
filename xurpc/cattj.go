//某个标签的实时数据
package main

import (
	"fmt"
	"github.com/ngaut/log"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"gopkg.in/olivere/elastic.v3"
)

type TaoCatData struct {
}

// 获取淘内的数据
func (this TaoCatData) GetTaoData(cid string) int {
	var (
		db    = IniFile.Section("mongo-data_source").Key("db").String()
		table = "useraction_put_big"
		sess  = getcattjSession()
	)
	defer sess.Close()

	c, err := sess.DB(db).C(table).Find(bson.M{"tag.tagId": cid}).Count()
	if err != nil {
		return 0
	}
	return c
}

// 获取医疗的数据
func (this TaoCatData) GetHospitalData(cid string) int {
	var (
		db    = IniFile.Section("mongo-data_source").Key("db").String()
		table = "useraction_put_big"
		sess  = getcattjSession()
	)
	defer sess.Close()

	if !bson.IsObjectIdHex(cid) {
		return 0
	}

	c, err := sess.DB(db).C(table).Find(bson.M{"tag.tagId": cid}).Count()
	if err != nil {
		return 0
	}
	return c
}

// 获取域名的数据
func (this TaoCatData) GetDomainData(cid string) int {
	var (
		db    = IniFile.Section("mongo-data_source").Key("db").String()
		table = "urltrack_put"
		sess  = getcattjSession()
	)
	defer sess.Close()

	c, err := sess.DB(db).C(table).Find(bson.M{"cids.id": cid}).Count()
	if err != nil {
		return 0
	}
	return c
}

// 获取店铺数据量
func (this TaoCatData) GetShopCount(shopid string, date string) int {
	var hosts = IniFile.Section("es").Key("host").Strings(",")
	client, err := elastic.NewClient(elastic.SetURL(hosts...))
	if err != nil {
		log.Error(err)
		return 0
	}

	var body = `{
		"query" : {
			"filtered" : {
				"filter" : {
					"range" : {
						"timestamp" : { "gte" : "` + date + `" }
					}
				},
				"query" : {
					"term" : {
						"shop" : "` + shopid + `"
					}
				}
			}
		}
	}`
	num, err := client.Count().Index("zhejiang_tb_shop_trace").Type("shop").BodyJson(body).Do()

	if err != nil {
		log.Error(err)
		return 0
	}
	return int(num)
}

//获取session
func getcattjSession() *mgo.Session {
	var (
		mouser = IniFile.Section("mongo-data_source").Key("user").String()
		mopwd  = IniFile.Section("mongo-data_source").Key("pwd").String()
		mohost = IniFile.Section("mongo-data_source").Key("host").String()
		moport = IniFile.Section("mongo-data_source").Key("port").String()
		modb   = IniFile.Section("mongo-data_source").Key("db").String()
		url    = fmt.Sprintf("%s:%s/%s", mohost, moport, modb)
	)
	if mouser != "" && modb != "" {
		url = fmt.Sprintf("%s:%s@%s", mouser, mopwd, url)
	}

	mdbsession, err := mgo.Dial(url)
	if err != nil {
		panic(err)
	}
	return mdbsession

}

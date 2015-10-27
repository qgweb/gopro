//某个标签的实时数据
package main

import (
	"fmt"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
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

	tid := bson.ObjectIdHex(cid)

	c, err := sess.DB(db).C(table).Find(bson.M{"tag.tagId": tid}).Count()
	if err != nil {
		return 0
	}
	return c
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

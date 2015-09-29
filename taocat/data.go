// grab taocat
package main

import (
	"fmt"
	"github.com/ngaut/log"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"math"
)

var (
	mdbsession *mgo.Session
	mo_user    = "xu"
	mo_pwd     = "123456"
	mo_host    = "192.168.1.199"
	mo_port    = "27017"
	mo_db      = "xu_precise"
	mo_table   = "tao_cat"
)

//获取mongo数据库链接
func GetSession() *mgo.Session {
	var (
		mouser = mo_user
		mopwd  = mo_pwd
		mohost = mo_host
		moport = mo_port
		modb   = mo_db
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

// 数据导入
func importData() {
	sess := GetSession()
	defer sess.Close()

	count, _ := sess.DB(mo_db).C("taocat").Count()
	pageSize := 1000
	pageCount := int(math.Ceil(float64(count) / float64(pageSize)))

	for i := 1; i <= pageCount; i++ {
		list := make([]map[string]interface{}, 0, pageSize)
		sess.DB(mo_db).C("taocat").Find(bson.M{}).
			Select(bson.M{"cid": 1, "count": 1, "unit": 1, "_id": 0}).
			Limit(pageSize).Skip((i - 1) * pageSize).All(&list)

		for _, v := range list {
			log.Info(v)
			sess.DB(mo_db).C(mo_table).Update(bson.M{"cid": v["cid"]}, bson.M{"$set": bson.M(v)})
		}
	}
	log.Warn(count, pageCount)
}

//
func main() {
	importData()
}

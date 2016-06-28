package models

import (
	"encoding/json"

	"github.com/qgweb/new/lib/mongodb"
)

type Order struct {
	Id         string `json:"oid" bson:"oid" form:"oid" comm:"oid"`
	Name       string `json:"name" bson:"name" form:"name" comm:"订单名称"`
	Price      string `json:"price" bson:"price" form:"price" comm:"出价"`
	Size       string `json:"size" bson:"size" form:"size" comm:"投放尺寸"`
	Btime      string `json:"btime" bson:"btime" form:"btime" comm:"投放开始时间"`
	Etime      string `json:"etime" bson:"etime" form:"etime" comm:"投放结束时间"`
	DayLimit   string `json:"day_limit" bson:"day_limit" form:"day_limit" comm:"日限额"`
	TotalLimit string `json:"total_limit" bson:"total_limit" form:"total_limit" comm:"总限额"`
	TimePoint  string `json:"time_point" bson:"time_point" form:"time_point" comm:"投放时间段"`
	Purl       string `json:"purl" bson:"purl" form:"purl" comm:"物料地址"`
	Surls      string `json:"surls" bson:"surls" form:"surls" comm:"目标投放地址集合"`
	Stats      string `json:"stats" bson:"stats" form:"stats" comm:"状态"`
}

func (this *Order) parse(info map[string]interface{}) Order {
	if bs, err := json.Marshal(&info); err == nil {
		var o Order
		if err := json.Unmarshal(bs, &o); err == nil {
			return o
		}
	}
	return Order{}
}

func (this *Order) uparse(info Order) map[string]interface{} {
	if bs, err := json.Marshal(&info); err == nil {
		var o map[string]interface{}
		if err := json.Unmarshal(bs, &o); err == nil {
			return o
		}
	}
	return nil
}

func (this *Order) Add(o Order) error {
	mgo, err := mdb.Get()
	if err != nil {
		return err
	}
	defer mgo.Close()
	o.Id = mongodb.GetObjectId()
	qconf := mongodb.MongodbQueryConf{}
	qconf.Db = "shrtb"
	qconf.Table = "order"
	qconf.Insert = []interface{}{&o}
	return mgo.Insert(qconf)
}

func (this *Order) List() ([]Order, error) {
	mgo, err := mdb.Get()
	if err != nil {
		return nil, err
	}
	defer mgo.Close()
	qconf := mongodb.MongodbQueryConf{}
	qconf.Db = "shrtb"
	qconf.Table = "order"
	qconf.Query = nil
	var list = make([]Order, 0, 20)
	err = mgo.Query(qconf, func(info map[string]interface{}) {
		o := this.parse(info)
		list = append(list, o)
	})
	return list, err
}

func (this *Order) Del(id string) error {
	mgo, err := mdb.Get()
	if err != nil {
		return err
	}
	defer mgo.Close()
	qconf := mongodb.MongodbQueryConf{}
	qconf.Db = "shrtb"
	qconf.Table = "order"
	qconf.Delete = mongodb.MM{"oid": (id)}
	return mgo.Delete(qconf)
}

func (this *Order) GetId(id string) (o Order, err error) {
	mgo, err := mdb.Get()
	if err != nil {
		return o, err
	}
	defer mgo.Close()
	qconf := mongodb.MongodbQueryConf{}
	qconf.Db = "shrtb"
	qconf.Table = "order"
	qconf.Query = mongodb.MM{"oid": (id)}
	if info, err := mgo.One(qconf); err == nil {
		return this.parse(info), nil
	} else {
		return o, err
	}
}

func (this *Order) Edit(o Order) error {
	mgo, err := mdb.Get()
	if err != nil {
		return err
	}
	defer mgo.Close()
	qconf := mongodb.MongodbQueryConf{}
	qconf.Db = "shrtb"
	qconf.Table = "order"
	qconf.Update = mongodb.MM{"$set": mongodb.MM(this.uparse(o))}
	qconf.Query = mongodb.MM{"oid": o.Id}
	//qconf.Update = mongodb.MM{"$set": mongodb.MM{"name": "bbbb"}}
	//qconf.Query = mongodb.MM{"oid": "577270b933f5247449e31724"}
	return mgo.Update(qconf)
}

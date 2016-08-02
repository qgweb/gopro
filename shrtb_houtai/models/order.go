package models

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/qgweb/new/lib/convert"

	"github.com/lisijie/cron"
	"github.com/qgweb/new/lib/mongodb"
	"github.com/qgweb/new/lib/timestamp"
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
	o.Stats = "未投放"
	if strings.TrimSpace(o.TotalLimit) == "" {
		o.TotalLimit = "9999999"
	}
	if strings.TrimSpace(o.DayLimit) == "" {
		o.DayLimit = "9999999"
	}
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
	if strings.TrimSpace(o.TotalLimit) == "" {
		o.TotalLimit = "9999999"
	}
	if strings.TrimSpace(o.DayLimit) == "" {
		o.DayLimit = "9999999"
	}
	qconf := mongodb.MongodbQueryConf{}
	qconf.Db = "shrtb"
	qconf.Table = "order"
	qconf.Update = mongodb.MM{"$set": mongodb.MM(this.uparse(o))}
	qconf.Query = mongodb.MM{"oid": o.Id}
	//qconf.Update = mongodb.MM{"$set": mongodb.MM{"name": "bbbb"}}
	//qconf.Query = mongodb.MM{"oid": "577270b933f5247449e31724"}
	return mgo.Update(qconf)
}

func (this *Order) getPushUrls(urls string) map[string]byte {
	var res = make(map[string]byte)
	for _, v := range strings.Split(urls, "\n") {
		v = strings.TrimSpace(v)
		if v != "" {
			res[v] = 1
		}
	}
	return res
}

// 推送
func (this *Order) Push(id string, status string) error {
	o, err := this.GetId(id)
	if err != nil {
		return err
	}
	puturl := fmt.Sprintf("%s\t%s\t%s\t%s", o.Price, o.Size, o.Purl, o.Id)
	if status == "1" {
		res, err := this.GetStatus(id)
		if err != nil {
			return err
		}
		o.Stats = res
		if err := this.Edit(o); err != nil {
			return err
		}
		if res == "投放中" {
			if status == "1" {
				for v := range this.getPushUrls(o.Surls) {
					putDb.Sadd(v, puturl)
				}
			}
		}
		if res == "未投放" {
			for v := range this.getPushUrls(o.Surls) {
				putDb.Srem(v, puturl)
			}
		}
	}

	if status == "0" {
		o.Stats = "未投放"
		if err := this.Edit(o); err != nil {
			return err
		}
		for v := range this.getPushUrls(o.Surls) {
			putDb.Srem(v, puturl)
		}
	}
	// 自定义
	if status == "self" {
		res, err := this.GetStatus(id)
		if err != nil {
			return err
		}
		o.Stats = res
		if err := this.Edit(o); err != nil {
			return err
		}
		for v := range this.getPushUrls(o.Surls) {
			putDb.Srem(v, puturl)
		}
	}

	return nil
}

func (this *Order) GetStatus(order_id string) (string, error) {
	var (
		r Report
		d = convert.ToInt64(timestamp.GetDayTimestamp(0))
	)

	o, err := this.GetId(order_id)
	if err != nil {
		return "未投放", err
	}

	rinfo, err := r.GetOne(map[string]interface{}{"order_id": order_id, "date": d})
	if err != nil && err.Error() != "not found" {
		return "未投放", err
	}

	btime := convert.ToInt64(timestamp.GetTimestamp(o.Btime + " 00:00:00"))
	etime := convert.ToInt64(timestamp.GetTimestamp(o.Etime + " 00:00:00"))
	tpv, err := r.GetTotalPv(map[string]interface{}{
		"order_id": order_id,
		"date":     map[string]interface{}{"$gte": btime, "$lte": etime},
	})

	if err != nil {
		return "未投放", err
	}

	if convert.ToInt64(o.DayLimit)*1000 < rinfo.Pv {
		return "日限额已到", nil
	}
	if convert.ToInt64(o.TotalLimit)*1000 < tpv {
		return "总限额已到", nil
	}
	if btime > d || etime < d {
		return "不在投放周期范围内", nil
	}
	if !this.CheckInWeekDay(o.TimePoint) {
		return "不在投放时间段内", nil
	}

	return "投放中", nil
}

func (this *Order) CheckInWeekDay(info string) bool {
	infos := strings.Split(info, "|")
	w := int(time.Now().Weekday())
	if w == 0 {
		w = 6
	} else {
		w = w - 1
	}

	h := time.Now().Hour()
	if len(infos) < w {
		return false
	}
	if hs := strings.Split(infos[w], ""); len(hs) >= h && hs[h] == "1" {
		return true
	}
	return false
}

func (this *Order) TimingCheckStats() {
	list, err := this.List()
	if err != nil {
		fmt.Println("[ERROR]", err)
		return
	}
	for _, order := range list {
		if order.Stats != "未投放" {
			stats, err := this.GetStatus(order.Id)
			if err != nil {
				continue
			}

			if stats != "投放中" {
				this.Push(order.Id, "self")
			} else {
				this.Push(order.Id, "1")
			}
		}
	}
}

func (this *Order) Putin() {
	c := cron.New()
	c.Start()
	c.AddFunc("*/30 * * * * *", func() {
		fmt.Println("ok")
		this.TimingCheckStats()
	})
}

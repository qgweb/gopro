package models

import (
	"goclass/convert"
	"time"

	"strings"

	"github.com/astaxie/beego"
	"github.com/qgweb/new/lib/mongodb"
	"github.com/qgweb/new/lib/timestamp"
)

type Report struct {
	Id      string `json:"rid" bson:"rid" comm:"编号"`
	OrderId string `json:"order_id" bson:"order_id" comm:"订单id"`
	Pv      int64  `json:"pv" bson:"pv" comm:"pv"`
	Date    int64  `json:"date" bson:"date" comm:"日期"`
	Time    int64  `json:"time" bson:"time" comm:"时间"`
}

type ReportEx struct {
	Report
	OrderName string `json:"order_name" bson:"order_name" comm:"订单名称"`
}

func (this *Report) Add(r Report) error {
	mb, err := mdb.Get()
	if err != nil {
		return err
	}
	defer mb.Close()
	r.Id = mongodb.GetObjectId()
	qconf := mongodb.MongodbQueryConf{}
	qconf.Db = "shrtb"
	qconf.Table = "report"
	qconf.Insert = []interface{}{&r}
	return mb.Insert(qconf)
}

func (this *Report) GetOne(query map[string]interface{}) (o Report, err error) {
	mgo, err := mdb.Get()
	if err != nil {
		return o, err
	}
	defer mgo.Close()
	qconf := mongodb.MongodbQueryConf{}
	qconf.Db = "shrtb"
	qconf.Table = "report"
	qconf.Query = mongodb.MM(query)
	if info, err := mgo.One(qconf); err == nil {
		var r Report
		Parse(info, &r)
		return r, nil
	} else {
		return o, err
	}
}

func (this *Report) GetMul(query map[string]interface{}) (list []Report, err error) {
	mgo, err := mdb.Get()
	if err != nil {
		return nil, err
	}
	defer mgo.Close()
	qconf := mongodb.MongodbQueryConf{}
	qconf.Db = "shrtb"
	qconf.Table = "report"
	qconf.Query = mongodb.MM(query)
	list = make([]Report, 0, 10)
	mgo.Query(qconf, func(info map[string]interface{}) {
		var r Report
		if Parse(info, &r) != nil {
			list = append(list, r)
		}
	})
	return list, nil
}

func (this *Report) IncrPv(oid string, date int64) error {
	mgo, err := mdb.Get()
	if err != nil {
		return err
	}
	defer mgo.Close()
	qconf := mongodb.MongodbQueryConf{}
	qconf.Db = "shrtb"
	qconf.Table = "report"
	qconf.Update = mongodb.MM{"$inc": mongodb.MM{"pv": 1}}
	qconf.Query = mongodb.MM{"order_id": oid, "date": date}
	return mgo.Update(qconf)
}

func (this *Report) GetTotalPv(query map[string]interface{}) (int64, error) {
	list, err := this.GetMul(query)
	if err != nil {
		return 0, err
	}
	var sum int64
	for _, v := range list {
		sum += v.Pv
	}
	return sum, nil
}

func (this *Report) List() ([]ReportEx, error) {
	mgo, err := mdb.Get()
	if err != nil {
		return nil, err
	}
	defer mgo.Close()
	qconf := mongodb.MongodbQueryConf{}
	qconf.Db = "shrtb"
	qconf.Table = "report"
	qconf.Query = nil
	var list = make([]ReportEx, 0, 20)
	var or Order
	err = mgo.Query(qconf, func(info map[string]interface{}) {
		var r ReportEx
		Parse(info, &r)
		o, _ := or.GetId(r.OrderId)
		r.OrderName = o.Name
		list = append(list, r)
	})
	return list, err
}

// 处理展现请求
func (this *Report) LoopPvStats() {
	for {
		d := convert.ToInt64(timestamp.GetDayTimestamp(0))
		//格式：mid
		omid, err := putDbLoop.Lpop("SHRTB_PV_QUEUE")

		if err != nil {
			time.Sleep(time.Second * 5)
			if strings.Contains(err.Error(), "close") || strings.Contains(err.Error(), "short") {
				putDbLoop = nil
				initPutDbLoop()
			}
			beego.Error(err)
			continue
		}

		mid := convert.ToString(omid)
		if omid != nil {
			beego.Info(mid)
		}

		if mid == "" {
			continue
		}

		var mkey = "MID_" + mid
		orderId := putDbLoop.Get(mkey)
		if orderId == "" {
			continue
		}

		_, err = this.GetOne(map[string]interface{}{
			"order_id": orderId,
			"date":     d,
		})
		if err != nil && err.Error() == "not found" {
			//创建
			this.Add(Report{
				OrderId: orderId,
				Pv:      1,
				Date:    d,
				Time:    time.Now().Unix(),
			})
		}
		if err == nil {
			//修改
			this.IncrPv(orderId, d)
		}
		//删除处理过的订单
		putDbLoop.Del(mkey)
		putDbLoop.Flush()
		putDbLoop.Flush()
	}
}

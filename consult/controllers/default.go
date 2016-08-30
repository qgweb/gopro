package controllers

import (
	"github.com/astaxie/beego"
	"gopkg.in/mgo.v2"
	"fmt"
	"github.com/ngaut/log"
	"github.com/qgweb/new/lib/timestamp"
	"github.com/qgweb/new/lib/convert"
	"gopkg.in/mgo.v2/bson"
	"sync"
)

var (
	actChan = make(chan byte, 100)
	monconn *mgo.Session
	msgChan = make(chan string, 1000)
	isChan bool
)

func init() {
	var (
		err error
		host = beego.AppConfig.String("mongodb::host")
		port = beego.AppConfig.String("mongodb::port")
		db = beego.AppConfig.String("mongodb::db")
	)
	monconn, err = mgo.Dial(fmt.Sprintf("mongodb://%s:%s/%s", host, port, db))
	if err != nil {
		log.Fatal(err)
	}
	monconn.SetMode(mgo.Monotonic, true)
}

type Infomation struct {
	Name     string `form:"name" bson:"name" json:"name"`
	Phone    string `form:"phone" bson:"phone" json:"phone"`
	QQ       string `form:"qq" bson:"qq" json:"qq"`
	Industry string `form:"industry" bson:"industry" json:"industry"`
	Date     int64 `form:"-" bson:"date" json:"date"`
	Reffer   string `form:"reffer" bson:"reffer" json:"reffer"`
}

type MainController struct {
	beego.Controller
}

func (c *MainController) JsonResult(code int, msg string, data interface{}) {
	c.Data["json"] = map[string]interface{}{
		"ret":code,
		"msg":msg,
		"data":data,
	}
	c.ServeJSON()
}

func (c *MainController) Get() {
	var o sync.Once
	c.TplName = "index.tpl"
	o.Do(func() {
		if !isChan {
			go c.Stats()
			isChan = true
		}
	})
}

func (c*MainController) StatsRec() {
	go func() {
		msgChan <- c.GetString("t")
	}()
	c.Ctx.WriteString("")
}

func (c *MainController) Stats() {
	var db = beego.AppConfig.String("mongodb::db")
	var sess = monconn.Clone()
	defer sess.Clone()

	for {
		select {
		case msg := <-msgChan:
			var dt = timestamp.GetDayTimestamp(0)
			if msg == "1" {
				log.Error(sess.DB(db).C("cstats").Upsert(bson.M{"date":dt},
					bson.M{"$inc":bson.M{"cs":1}}))
			}
			if msg == "2" {
				log.Error(sess.DB(db).C("cstats").Upsert(bson.M{"date":dt},
					bson.M{"$inc":bson.M{"fm":1}}))
			}
			if msg == "3" {
				log.Error(sess.DB(db).C("cstats").Upsert(bson.M{"date":dt},
					bson.M{"$inc":bson.M{"pv":1}}))
			}
		}
	}
}

func (c *MainController) Submit() {
	actChan <- 1
	defer func() {
		<-actChan
	}()

	var db = beego.AppConfig.String("mongodb::db")
	var sess = monconn.Clone()
	defer sess.Close()

	if c.Ctx.Input.IsPost() {
		var info Infomation
		if err := c.ParseForm(&info); err != nil {
			c.Ctx.WriteString("hehe")
			return;
		}
		info.Date = convert.ToInt64(timestamp.GetTimestamp())
		if err := sess.DB(db).C("info").Insert(info); err != nil {
			c.JsonResult(-1, "提交失败", nil)
			return
		}

		c.JsonResult(0, "提交成功，感谢您的预约", nil)
	}
}

func (c *MainController) List() {
	var (
		db = beego.AppConfig.String("mongodb::db")
		sess = monconn.Clone()
		list []Infomation
		list2 []map[string]interface{}
	)
	defer sess.Clone()
	err := sess.DB(db).C("info").Find(nil).Sort("date:-1").All(&list)
	if err != nil {
		beego.Error(err)
		c.Ctx.WriteString("读取数据出错")
		return
	}
	err = sess.DB(db).C("cstats").Find(nil).Sort("date:-1").All(&list2)
	if err != nil {
		beego.Error(err)
		c.Ctx.WriteString("读取数据出错")
		return
	}
	c.Data["list1"] = list
	c.Data["list2"] = list2
	c.TplName = "list.tpl"
}
package controllers

import (
	"github.com/astaxie/beego"
	"github.com/ngaut/log"
	"github.com/qgweb/gopro/lib/ssh"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type MainController struct {
	beego.Controller
}

func (c *MainController) Get() {
	conf := ssh.Config{}
	conf.PrivaryKey = ssh.GetPrivateKey()
	conf.RemoteHost = "122.225.98.69"
	conf.RemoteUser = "root"
	l, err := ssh.NewSSHLinker(conf)
	if err != nil {
		c.Ctx.WriteString(err.Error())
		return
	}
	s, err := l.GetClient().NewSession()
	if err != nil {
		log.Error(err)
		return
	}
	bb, err := s.Output("curl 'http://192.168.0.72:4151/stats'")
	if err != nil {
		log.Error(err)
		return
	}
	c.Ctx.WriteString(string(bb))
}

func (c *MainController) Mgo() {
	sess, err := mgo.Dial("xu:123456@192.168.1.199:27017/xu_precise")
	if err != nil {
		c.Ctx.WriteString(err.Error())
		return
	}
	// i := sess.DB("xu_precise").C("taocat").Find(bson.M{}).Iter()
	// var r map[string]interface{} = make(map[string]interface{})

	// for i.Next(&r) {
	// 	//c.Ctx.WriteString(r["name"].(string))
	// }

	// if err := i.Close(); err != nil {
	// 	c.Ctx.WriteString(err.Error())
	// }

	var list []map[string]interface{}
	sess.DB("xu_precise").C("taocat").Find(bson.M{}).All(&list)
	// for _, v := range list {
	// 	c.Ctx.WriteString(v["name"].(string))
	// }

	c.Ctx.WriteString("xxx")
}

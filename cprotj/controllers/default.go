package controllers

import (
	"github.com/astaxie/beego"
	"github.com/garyburd/redigo/redis"
	"github.com/juju/errors"
	"github.com/qgweb/gopro/cprotj/common/mongo"
	"github.com/qgweb/gopro/cprotj/common/xredis"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"time"
)

type MainController struct {
	beego.Controller
}

func (c *MainController) Error(err ...error) {
	beego.Error(err)
	c.Ctx.Output.SetStatus(500)
	c.Ctx.Output.Body([]byte("program error"))
}

func (c *MainController) pushAdToReids(conn redis.Conn, ad string) {
	var (
		db  = xredis.GetConfig().Db
		key = time.Now().Format("20060102") + "_cookie"
	)
	conn.Do("SELECT", db)
	conn.Send("HINCRBY", key, ad, 1)
	conn.Flush()
}

func (c *MainController) checkInUserCookies(conn *mgo.Session, uid string) (string, error) {
	var (
		db    = mongo.GetConfig().Db
		table = "dt_user"
	)

	if !bson.IsObjectIdHex(uid) {
		return "", errors.New(uid + ": is not a mongoid")
	}

	var info map[string]interface{}
	err := conn.DB(db).C(table).FindId(bson.ObjectIdHex(uid)).One(&info)
	if err != nil {
		return "", err
	}

	if v, ok := info["cox"]; ok {
		return v.(string), nil
	}
	return "", errors.New("cox 不存在")
}

func (c *MainController) CookieMatch() {
	c.Ctx.WriteString("ok")
	return
	var (
		rconn      = xredis.GetRedis()
		mconn, err = mongo.LinkMongo()
		uid        = c.Ctx.Input.Cookie("dt_uid")
	)

	defer func() {
		rconn.Close()
		mconn.Close()
	}()

	if err != nil || rconn.Err() != nil {
		c.Error(err, rconn.Err())
		return
	}

	ad, err := c.checkInUserCookies(mconn, uid)
	if err != nil {
		c.Error(err)
		return
	}

	if ad != "" {
		c.pushAdToReids(rconn, ad)
	}

	c.Ctx.WriteString("ok")
}

func (c *MainController) Reffer() {
	var (
		tu  = c.GetString("tu", "")
		ref = c.Ctx.Input.Header("Referer")
	)
	beego.Info(tu, "---------", ref)
	c.Ctx.WriteString("ok")
}

func (c *MainController) Iframe() {
	beego.Info(c.GetString("if"))
}
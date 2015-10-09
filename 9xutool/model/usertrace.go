package model

import (
	"io/ioutil"
	"time"

	"github.com/codegangsta/cli"
	"github.com/ngaut/log"
	"github.com/qgweb/gopro/9xutool/common.go"
	"github.com/qgweb/gopro/lib/convert"
	"gopkg.in/ini.v1"
	"gopkg.in/mgo.v2/bson"
)

type UserTrace struct {
	mp      *common.MgoPool
	iniFile *ini.File
}

func NewUserTraceCli() cli.Command {
	return cli.Command{
		Name:    "user_trace_merge",
		Aliases: []string{"a"},
		Usage:   "生成用户最近3天浏览轨迹,供九旭精准投放",
		Action: func(c *cli.Context) {
			defer func() {
				if msg := recover(); msg != nil {
					log.Error(msg)
				}
			}()

			// 获取配置文件
			filePath := common.GetBasePath() + "/conf/ut.conf"
			f, err := ioutil.ReadFile(filePath)
			if err != nil {
				log.Fatal(err)
				return
			}

			ur := &UserTrace{}
			ur.iniFile, _ = ini.Load(f)

			// mgo 配置文件
			mconf := &common.MgoConfig{}
			mconf.DBName = ur.iniFile.Section("mongo").Key("db").String()
			mconf.Host = ur.iniFile.Section("mongo").Key("host").String()
			mconf.Port = ur.iniFile.Section("mongo").Key("port").String()
			mconf.UserName = ur.iniFile.Section("mongo").Key("user").String()
			mconf.UserPwd = ur.iniFile.Section("mongo").Key("pwd").String()
			ur.mp = common.NewMgoPool(mconf)

			ur.Do(c)
		},
		//		Flags: []cli.Flag{
		//			cli.StringFlag{"port, p", "3000", "Temporary port number to prevent conflict", ""},
		//			cli.StringFlag{"config, c", "custom/conf/app.ini", "Custom configuration file path", ""},
		//		},
	}
}

func (this *UserTrace) Do(c *cli.Context) {
	var (
		date  = time.Now()
		day   = date.Format("20060102")
		hour  = convert.ToString(date.Hour() - 1)
		b1day = date.AddDate(0, 0, -1).Format("20060102") //1天前
		b2day = date.AddDate(0, 0, -2).Format("20060102") //2天前
		b3day = date.AddDate(0, 0, -3).Format("20060102") //3天前
		sess  = this.mp.Get()
		db    = this.iniFile.Section("mongo").Key("db").String()
		table = "useraction"
		list  []map[string]interface{}
		list1 []map[string]interface{}
		list2 []map[string]interface{}
		list3 []map[string]interface{}
	)

	// 当天前一个小时前的数据
	if err := sess.DB(db).C(table).Find(bson.M{"day": day, "hour": bson.M{
		"$lte": hour, "$gte": "00"},
	}).All(&list1); err != nil {
		log.Error(err)
	}

	// 前2天数据
	if err := sess.DB(db).C(table).Find(bson.M{"day": bson.M{
		"$lte": b1day, "$gte": b2day},
	}).All(&list2); err != nil {
		log.Error(err)
	}

	// 第前3天的小时数据
	if err := sess.DB(db).C(table).Find(bson.M{"day": b3day, "hour": bson.M{
		"$lte": hour, "$gte": "23"},
	}).All(&list3); err != nil {
		log.Error(err)
	}

	appendFun := func(l []map[string]interface{}) {
		for _, v := range l {
			for _, v1 := range list {
				if v["AD"] == v1["AD"] && v["UA"] == v1["UA"] {
					if _, ok := v["tags"]; !ok {
						v["tags"] = make([]bson.M, 0, 10)
					}

					// 去重
					for _, v2 := range v["tag"].([]bson.M) {
						ise := false
						for _, v3 := range v1["tag"].([]bson.M) {
							if v2["tagId"] == v3["tagId"] {
								ise = true
							}
						}

						if !ise {
							v1["tag"] = append(v1["tag"].([]bson.M), v2)
						}
					}
				}
			}
		}
	}

	// 合并数据
	appendFun(list1)
	appendFun(list2)
	appendFun(list3)

	log.Info(b2day, day, b3day)
	log.Info(list)
	sess.Close()
}

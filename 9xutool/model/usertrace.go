package model

import (
	"io/ioutil"
	"math"
	"strings"
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
		Name:  "user_trace_merge",
		Usage: "生成用户最近3天浏览轨迹,供九旭精准投放",
		Action: func(c *cli.Context) {
			defer func() {
				//				if msg := recover(); msg != nil {
				//					log.Error(msg)
				//				}
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
		date      = time.Now()
		day       = date.Format("20060102")
		hour      = convert.ToString(date.Hour() - 1)
		b1day     = date.AddDate(0, 0, -1).Format("20060102") //1天前
		b2day     = date.AddDate(0, 0, -2).Format("20060102") //2天前
		b3day     = date.AddDate(0, 0, -3).Format("20060102") //3天前
		sess      = this.mp.Get()
		db        = this.iniFile.Section("mongo").Key("db").String()
		table     = "useraction"
		table_put = "useraction_put"
		list_put  []interface{}
		list      map[string][]map[string]interface{}
		list1     []map[string]interface{}
		list2     []map[string]interface{}
		list3     []map[string]interface{}
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
		"$gte": hour, "$lte": "23"},
	}).All(&list3); err != nil {
		log.Error(err)
	}

	var appendFun = func(l []map[string]interface{}) {
		for _, v := range l {
			key := v["AD"].(string) + "_" + v["UA"].(string)
			if tag, ok := list[key]; ok {
				//去重
				for _, tv := range v["tag"].([]interface{}) {
					isee := false
					tvm := tv.(map[string]interface{})
					for _, tv1 := range tag {
						if tvm["tagId"] == tv1["tagId"] {
							isee = true
							break
						}
					}
					if !isee {
						list[key] = append(list[key], tvm)
					}
				}
			} else {
				tag := make([]map[string]interface{}, 0, len(v["tag"].([]interface{})))
				for _, vv := range v["tag"].([]interface{}) {
					tag = append(tag, vv.(map[string]interface{}))
				}
				list[key] = tag
			}
		}
	}

	// 合并数据
	list = make(map[string][]map[string]interface{})
	appendFun(list1)
	appendFun(list2)
	appendFun(list3)

	//更新投放表
	list_put = make([]interface{}, 0, len(list))
	for k, v := range list {
		uaads := strings.Split(k, "_")
		list_put = append(list_put, bson.M{
			"AD":  uaads[0],
			"UA":  uaads[1],
			"tag": v,
		})
	}

	sess.DB(db).C(table_put).DropCollection()
	//批量插入
	var (
		size  = 10000
		count = int(math.Ceil(float64(len(list_put)) / float64(size)))
	)

	for i := 1; i <= count; i++ {
		end := (i-1)*size + size
		if end > len(list_put) {
			end = len(list_put)
		}

		sess.DB(db).C(table_put).Insert(list_put[(i-1)*size : end]...)
		log.Error(i, size)
	}

	sess.Close()
	log.Info("ok")
}

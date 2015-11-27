package model

import (
	"github.com/qgweb/gopro/lib/encrypt"
	"io/ioutil"
	"math"
	"strings"
	"time"

	"github.com/codegangsta/cli"
	"github.com/ngaut/log"
	"github.com/qgweb/gopro/9xutool/common"
	"gopkg.in/ini.v1"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type URLTrace struct {
	mp      *common.MgoPool
	iniFile *ini.File
}

func NewURLTraceCli() cli.Command {
	return cli.Command{
		Name:  "url_trace_merge",
		Usage: "生成域名1天的数据,供九旭精准投放",
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

			ur := &URLTrace{}
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
	}
}

func (this *URLTrace) Do(c *cli.Context) {
	var (
		date      = time.Now()
		day       = date.Format("20060102")
		b1day     = date.AddDate(0, 0, -1).Format("20060102") //1天前
		hour      = date.Add(-time.Hour).Format("15")
		month     = time.Now().Format("200601")
		table     = "urltrack_" + month
		table_put = "urltrack_put"
		sess      = this.mp.Get()
		db        = this.iniFile.Section("mongo").Key("db").String()
		list      map[string][]map[string]interface{}
	)
	defer sess.Close()

	list = make(map[string][]map[string]interface{})
	var appendFun = func(l []map[string]interface{}) {
		for _, v := range l {
			ad := v["ad"].(string)
			ua := "ua"
			if u, ok := v["ua"]; ok {
				ua = u.(string)
			}

			key := ad + "_" + ua
			if tag, ok := list[key]; ok {
				//去重
				for _, tv := range v["cids"].([]interface{}) {
					isee := false
					tvm := tv.(map[string]interface{})
					for _, tv1 := range tag {
						if tvm["id"] == tv1["id"] {
							isee = true
							break
						}
					}
					if !isee {
						list[key] = append(list[key], tvm)
					}
				}
			} else {
				tag := make([]map[string]interface{}, 0, len(v["cids"].([]interface{})))
				for _, vv := range v["cids"].([]interface{}) {
					tag = append(tag, vv.(map[string]interface{}))
				}
				list[key] = tag
			}
		}
	}
	// 读取数据函数
	var readDataFun = func(query ...interface{}) {
		var (
			count     = 0
			page      = 1
			pageSize  = 100000
			totalPage = 0
			sess      = this.mp.Get()
			querym    bson.M
		)

		if v, ok := query[0].(bson.M); ok {
			querym = v
		} else {
			return
		}

		count, err := sess.DB(db).C(table).Find(querym).Count()
		if err != nil {
			log.Error(err)
			return
		}

		totalPage = int(math.Ceil(float64(count) / float64(pageSize)))

		for ; page <= totalPage; page++ {
			var tmpList []map[string]interface{}
			if err := sess.DB(db).C(table).Find(querym).
				Select(bson.M{"_id": 0, "ad": 1, "ua": 1, "cids": 1}).
				Limit(pageSize).
				Skip((page - 1) * pageSize).All(&tmpList); err != nil {
				log.Error(err)
				continue
			}

			appendFun(tmpList)
			log.Warn(len(list))
		}

		sess.Close()
	}

	//读取数据
	_ = b1day
	//_ = bson.M{"date": day, "hour": bson.M{"$lte": hour, "$gte": hour}}
	readDataFun(bson.M{"date": day, "hour": bson.M{"$lte": hour, "$gte": hour}})
	//readDataFun(bson.M{"date": b1day, "hour": bson.M{"$lte": "23", "$gte": hour}})

	//更新投放表
	log.Info(len(list))
	sess.DB(db).C(table_put).DropCollection()

	//加索引
	sess.DB(db).C(table_put).Create(&mgo.CollectionInfo{})
	sess.DB(db).C(table_put).EnsureIndexKey("cids.id")
	sess.DB(db).C(table_put).EnsureIndexKey("adua")

	var (
		size     = 10000
		list_num = len(list)
	)

	list_put := make([]interface{}, 0, size)
	for k, v := range list {
		adua := strings.Split(k, "_")
		list_put = append(list_put, bson.M{
			"ad":   adua[0],
			"ua":   adua[1],
			"adua": encrypt.DefaultMd5.Encode(adua[0] + adua[1]),
			"cids": v,
		})

		if len(list_put) == size || len(list_put) == list_num {
			sess.DB(db).C(table_put).Insert(list_put...)
			log.Warn(len(list_put))
			list_put = make([]interface{}, 0, size)
			list_num = list_num - size
		}
	}
}

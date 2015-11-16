package model

import (
	"io/ioutil"
	"math"
	"strings"
	"time"

	"github.com/codegangsta/cli"
	"github.com/ngaut/log"
	"github.com/qgweb/gopro/9xutool/common"
	"github.com/qgweb/gopro/lib/encrypt"
	"gopkg.in/ini.v1"
	"gopkg.in/mgo.v2/bson"
)

type ShopTrace struct {
	mp      *common.MgoPool
	mpput   *common.MgoPool
	iniFile *ini.File
}

func NewShopTraceCli() cli.Command {
	return cli.Command{
		Name:  "shop_trace_merge",
		Usage: "生成用户最近3天店铺轨迹,供九旭精准投放",
		Action: func(c *cli.Context) {
			defer func() {
				if msg := recover(); msg != nil {
					log.Error(msg)
				}
			}()

			// 获取配置文件
			filePath := common.GetBasePath() + "/conf/st.conf"
			f, err := ioutil.ReadFile(filePath)
			if err != nil {
				log.Fatal(err)
				return
			}

			ur := &ShopTrace{}
			ur.iniFile, _ = ini.Load(f)

			// mgo 配置文件
			mconf := &common.MgoConfig{}
			mconf.DBName = ur.iniFile.Section("mongo").Key("db").String()
			mconf.Host = ur.iniFile.Section("mongo").Key("host").String()
			mconf.Port = ur.iniFile.Section("mongo").Key("port").String()
			mconf.UserName = ur.iniFile.Section("mongo").Key("user").String()
			mconf.UserPwd = ur.iniFile.Section("mongo").Key("pwd").String()
			ur.mp = common.NewMgoPool(mconf)

			mconfput := &common.MgoConfig{}
			mconfput.DBName = ur.iniFile.Section("mongo-put").Key("db").String()
			mconfput.Host = ur.iniFile.Section("mongo-put").Key("host").String()
			mconfput.Port = ur.iniFile.Section("mongo-put").Key("port").String()
			mconfput.UserName = ur.iniFile.Section("mongo-put").Key("user").String()
			mconfput.UserPwd = ur.iniFile.Section("mongo-put").Key("pwd").String()
			ur.mpput = common.NewMgoPool(mconfput)

			ur.Do(c)
		},
	}
}

func (this *ShopTrace) Do(c *cli.Context) {
	var (
		ndate     = time.Now()
		btime     = ndate.Format("2006-01-02")
		etime     = ndate.AddDate(0, 0, -3).Format("2006-01-02")
		sess      = this.mp.Get()
		sessput   = this.mpput.Get()
		db        = this.iniFile.Section("mongo").Key("db").String()
		dbput     = this.iniFile.Section("mongo-put").Key("db").String()
		count     = 0
		pageSize  = 10000
		maplist   = make(map[string][]string)
		putlist   []interface{}
		table_put = "useraction_shop"
	)

	defer sess.Close()
	defer sessput.Close()

	count, err := sess.DB(db).C("zhejiang_ad_tags_shop").Find(bson.M{"date": bson.M{"$lte": btime, "$gte": etime}}).Count()
	if err != nil {
		log.Error(err)
		return
	}

	pageCount := int(math.Ceil(float64(count) / float64(pageSize)))

	for i := 1; i <= pageCount; i++ {
		var list []map[string]interface{}
		err := sess.DB(db).C("zhejiang_ad_tags_shop").Find(bson.M{"date": bson.M{"$lte": btime, "$gte": etime}}).
			Skip((i - 1) * pageSize).Limit(pageSize).Sort("-date").All(&list)
		if err != nil {
			log.Error(err)
			continue
		}

		if len(list) <= 0 {
			continue
		}

		for _, v := range list {
			ua := v["ua"].(string)
			ad := v["ad"].(string)
			key := ad + "_" + ua
			if _, ok := maplist[ad+ua]; !ok {
				maplist[key] = make([]string, 0, 20)
			}

			if sp, ok := v["shop"].([]interface{}); ok {
				for _, spv := range sp {
					spvm := spv.(map[string]interface{})
					if spvm["id"].(string) != "" {
						maplist[key] = append(maplist[key], spvm["id"].(string))
					}
				}
			}
		}
	}

	//更新投放表
	putlist = make([]interface{}, 0, len(maplist))
	for k, v := range maplist {
		uaads := strings.Split(k, "_")
		putlist = append(putlist, bson.M{
			"ad":   uaads[0],
			"ua":   encrypt.DefaultBase64.Encode(uaads[1]),
			"shop": v,
		})
	}

	sessput.DB(dbput).C(table_put).DropCollection()
	//批量插入
	size := 10000
	count = int(math.Ceil(float64(len(putlist)) / float64(size)))

	for i := 1; i <= count; i++ {
		end := (i-1)*size + size
		if end > len(putlist) {
			end = len(putlist)
		}

		sessput.DB(dbput).C(table_put).Insert(putlist[(i-1)*size : end]...)
		log.Error(i, size)
	}
}

package model

import (
	"github.com/qgweb/gopro/lib/convert"
	"github.com/qgweb/gopro/lib/encrypt"
	"github.com/qgweb/gopro/lib/grab"
	"io/ioutil"
	"runtime/debug"
	"strings"
	"time"

	"github.com/codegangsta/cli"
	"github.com/ngaut/log"
	"github.com/qgweb/gopro/9xutool/common"
	"gopkg.in/ini.v1"
	"gopkg.in/mgo.v2/bson"
)

type TAGSTrace struct {
	mp      *common.MgoPool
	iniFile *ini.File
}

func NewTAGSTraceCli() cli.Command {
	return cli.Command{
		Name:  "get_tags_by_cookie",
		Usage: "根据用户ua和ad汇总十天内的标签",
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
				debug.PrintStack()
				return
			}

			ur := &TAGSTrace{}
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

func (this *TAGSTrace) Do(c *cli.Context) {
	var (
		timeLimit = getDayTime()
		table     = "useraction"
		table_put = "useraction_temp_tags"
		sess      = this.mp.Get()
		db        = this.iniFile.Section("mongo").Key("db").String()
	)
	defer sess.Close()

	//检查是否有数据,有就先清空
	dataCount, err := sess.DB(db).C(table_put).FindId(nil).Count()
	if err != nil {
		log.Error(err)
	}
	if dataCount == 0 {
		sess.DB(db).C(table_put).DropCollection()
	}

	//查询数据先
	var uaad []map[string]interface{}
	err = sess.DB(db).C(table).Find(bson.M{"timestamp": bson.M{"$gte": timeLimit}}).All(&uaad)
	if err != nil {
		log.Error(err)
	}

	var list map[string]map[string]int
	list = make(map[string]map[string]int)
	for _, v := range uaad {
		UA := "ua" //可能会没有UA
		if u, ok := v["UA"]; ok {
			UA = u.(string)
		}
		key := v["AD"].(string) + "_" + UA
		for _, tag := range v["tag"].([]interface{}) {
			tagm := tag.(map[string]interface{})
			if _, ok := list[key]; ok { //如果已经有这个用户相关的tag
				if _, ok := list[key][tagm["tagId"].(string)]; ok { //去重,判断是否已存在
					continue
				} else {
					list[key][tagm["tagId"].(string)] = convert.ToInt(v["timestamp"].(string))
				}
			} else {
				list[key] = map[string]int{tagm["tagId"].(string): convert.ToInt(v["timestamp"].(string))}
			}
		}
	}

	var (
		size     = 3
		list_num = len(list)
	)

	list_put := make([]interface{}, 0, size)

	for key, value := range list {
		tags_data := make([]string, 0, 5)
		tmp := make(grab.MapSorter, 0, 5)
		adua := strings.Split(key, "_")
		s := grab.NewMapSorter(value)
		s.Sort() //排序

		if count := len(s); count < 6 { //取前五个
			tmp = s[:count]
		} else {
			tmp = s[:5]
		}
		//查询出标签的中文名
		var result map[string]interface{}
		for _, r := range tmp {
			if bson.IsObjectIdHex(r.Key) {
				table = "category"
				err := sess.DB(db).C(table).FindId(bson.ObjectIdHex(r.Key)).One(&result)
				if err != nil {
					log.Error(err)
				}
			} else {
				table = "taocat"
				err := sess.DB(db).C(table).Find(bson.M{"cid": r.Key}).One(&result)
				if err != nil {
					log.Error(err)
				}
			}
			tags_data = append(tags_data, result["name"].(string))
		}

		list_put = append(list_put, bson.M{
			"ad":   adua[0],
			"ua":   adua[1],
			"adua": encrypt.DefaultMd5.Encode(key),
			"tag":  tags_data,
		})

		if len(list_put) == size || len(list_put) == list_num {
			sess.DB(db).C(table_put).Insert(list_put...)
			log.Warn(len(list_put))
			list_put = make([]interface{}, 0, size)
			list_num = list_num - size
		}
	}
	log.Info("ok")
}

/**
 * 获取十天前零点时间戳
 */
func getDayTime() string {
	d := time.Now().AddDate(0, 0, -10).Format("20060102")
	a, _ := time.ParseInLocation("20060102", d, time.Local)
	return convert.ToString(a.Unix())
}

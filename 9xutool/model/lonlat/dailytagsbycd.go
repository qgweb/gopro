package lonlat

import (
	"github.com/ngaut/log"
	"github.com/qgweb/gopro/lib/convert"
	"github.com/qgweb/gopro/lib/orm"
	"io/ioutil"
	"runtime/debug"

	"github.com/codegangsta/cli"
	"github.com/qgweb/gopro/9xutool/common"
	"gopkg.in/ini.v1"
	"gopkg.in/mgo.v2/bson"
)

type (
	TagInfo struct { //数据模型
		tagid string
		lon   string
		lat   string
		num   int
	}
)

var (
	tagsByJwd map[string]*TagInfo
)

func NewTagsCdCli() cli.Command {
	return cli.Command{
		Name:  "daily_tags_by_cd",
		Usage: "根据经纬度和ad汇总标签分布和标签经纬度统计",
		Action: func(c *cli.Context) {
			defer func() {
				if msg := recover(); msg != nil {
					log.Error(msg)
					debug.PrintStack()
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

			uc := &UserCdTrace{}
			uc.iniFile, _ = ini.Load(f)

			// mgo 配置文件
			mconf := &common.MgoConfig{}
			mconf.DBName = uc.iniFile.Section("mongo").Key("db").String()
			mconf.Host = uc.iniFile.Section("mongo").Key("host").String()
			mconf.Port = uc.iniFile.Section("mongo").Key("port").String()
			mconf.UserName = uc.iniFile.Section("mongo").Key("user").String()
			mconf.UserPwd = uc.iniFile.Section("mongo").Key("pwd").String()
			uc.mp = common.NewMgoPool(mconf)
			//mysql 配置文件
			uc.mysql = orm.NewORM()

			uc.initData()
			uc.Doit(c)
		},
	}
}

func (this *UserCdTrace) Doit(c *cli.Context) {
	var (
		db   = this.iniFile.Section("mongo").Key("db").String()
		sess = this.mp.Get()
	)
	iter := sess.DB(db).C(JWD_TABLE).Find(nil).Iter() //昨天的数据
	i := 1
	for {
		log.Info("已处理", convert.ToString(i), "条记录")
		i++
		var info map[string]interface{}
		if !iter.Next(&info) {
			break
		}
		this.getTagsJWDInfo(info)
	}

	if len(tagsByJwd) > 0 {
		this.getMysqlConnect() //连接mysql

		DayTimestamp := common.GetDayTimestamp(-1)
		for _, t := range tagsByJwd {
			if _, ok := taocat_list[t.tagid]; !ok {
				continue
			}
			this.mysql.BSQL().Insert("tags_report_jw").Values("tag_id", "lon", "lat", "num", "time")
			_, err := this.mysql.Insert(t.tagid, t.lon, t.lat, t.num, DayTimestamp)
			if err != nil {
				log.Warn("插入失败 ", err)
			}
		}
	}
	log.Info("数据分析完毕!")
}

//根据ad获取标签
func (this *UserCdTrace) getTagsJWDInfo(info map[string]interface{}) {
	var (
		db       = this.iniFile.Section("mongo").Key("db").String()
		sess     = this.mp.Get()
		tagsInfo []map[string]interface{}
		// dayTime  = getDay(-1) //0为今日
		// ad       = info["ad"].(string)
	)
	defer sess.Close()

	// err := sess.DB(db).C("useraction").Find(bson.M{"AD": ad, "day": dayTime}).All(&tagsInfo)
	err := sess.DB(db).C(USERACTION_TABLE).Find(bson.M{"AD": "YwdLb0cZUVlABmVXcAhgeg==", "day": "20151206"}).All(&tagsInfo)
	if err != nil {
		log.Error(err)
	}
	if len(tagsInfo) > 0 {
		tagsByJwd = make(map[string]*TagInfo)
		for _, v := range tagsInfo { //可能会有多条数据，即多个ad
			for _, tag := range v["tag"].([]interface{}) { //获取每个ad内的tagid
				tagm := tag.(map[string]interface{})
				if tagm["tagmongo"].(string) == "1" { //如果是mongoid忽略
					continue
				}
				cid := tagm["tagId"].(string)
				cg := taocat_list[cid] //从总标签的map判断是否是3级标签
				if cg.Level != 3 {
					cid, err = cg.getLv3Id()
					if err != nil {
						continue
					}
				}
				key := cid + info["lon"].(string) + info["lat"].(string)
				if t, ok := tagsByJwd[key]; ok {
					t.num++
					continue
				}
				tmp_info := &TagInfo{
					tagid: cid,
					lon:   info["lon"].(string),
					lat:   info["lat"].(string),
					num:   1,
				}
				tagsByJwd[key] = tmp_info
			}
		}
	}
}

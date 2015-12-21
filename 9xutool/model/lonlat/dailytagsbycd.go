package lonlat

import (
	"github.com/ngaut/log"
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
			filePath := common.GetBasePath() + "/conf/jw.conf"
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
			mconf.DBName = uc.iniFile.Section("mongo-data_source").Key("db").String()
			mconf.Host = uc.iniFile.Section("mongo-data_source").Key("host").String()
			mconf.Port = uc.iniFile.Section("mongo-data_source").Key("port").String()
			mconf.UserName = uc.iniFile.Section("mongo-data_source").Key("user").String()
			mconf.UserPwd = uc.iniFile.Section("mongo-data_source").Key("pwd").String()
			uc.debug, _ = uc.iniFile.Section("mongo-data_source").Key("debug").Int()
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
		db   = this.iniFile.Section("mongo-lonlat_data").Key("db").String()
		sess = this.mp.Get()
	)
	iter := sess.DB(db).C(JWD_TABLE).Find(nil).Iter()
	i := 1
	for {
		var info map[string]interface{}
		if !iter.Next(&info) {
			break
		}
		this.getTagsJWDInfo(info)
		if this.debug == 1 {
			log.Info("已处理", i, "条记录")
		}
		i++
	}

	if len(this.tagsByJwd) > 0 {
		this.getMysqlConnect() //连接mysql

		DayTimestamp := common.GetDayTimestamp(-1)
		for _, t := range this.tagsByJwd {
			if _, ok := this.taocat_list[t.tagid]; !ok {
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
		db        = this.iniFile.Section("mongo-data_source").Key("db").String()
		sess      = this.mp.Get()
		timestamp = common.GetDayTimestamp(-1) //0为今日
		ad        = info["AD"].(string)
		err       error
	)
	defer sess.Close()
	iter := sess.DB(db).C(USERACTION_TABLE).Find(bson.M{"AD": ad, "timestamp": timestamp}).Iter()
	// iter := sess.DB(db).C(USERACTION_TABLE).Find(bson.M{"AD": "YwdLb0cZUVlABmVXcAhgeg==", "day": "20151206"}).Iter()

	this.tagsByJwd = make(map[string]*TagInfo)
	for {
		var tagsInfo map[string]interface{}
		if !iter.Next(&tagsInfo) {
			break
		}
		for _, tag := range tagsInfo["tag"].([]interface{}) { //获取每个ad内的tagid
			tagm := tag.(map[string]interface{})
			if tagm["tagmongo"].(string) == "1" { //如果是mongoid忽略
				continue
			}
			cid := tagm["tagId"].(string)
			cg := this.taocat_list[cid] //从总标签的map判断是否是3级标签
			if cg.Level != 3 {
				cid, err = cg.getLv3Id(this)
				if err != nil {
					continue
				}
			}
			key := cid + info["Lon"].(string) + info["Lat"].(string)
			if t, ok := this.tagsByJwd[key]; ok {
				t.num++
				continue
			}
			tmp_info := &TagInfo{
				tagid: cid,
				lon:   info["Lon"].(string),
				lat:   info["Lat"].(string),
				num:   1,
			}
			this.tagsByJwd[key] = tmp_info
		}
	}
}

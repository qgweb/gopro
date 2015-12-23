package lonlat

import (
	"github.com/ngaut/log"
	"github.com/qgweb/gopro/lib/orm"
	"gopkg.in/mgo.v2"
	"io/ioutil"
	"runtime/debug"

	"github.com/codegangsta/cli"
	"github.com/qgweb/gopro/9xutool/common"
	"gopkg.in/ini.v1"
	"gopkg.in/mgo.v2/bson"
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

			// mgo 配置文件 经纬
			mconf_jw := &common.MgoConfig{}
			mconf_jw.DBName = uc.iniFile.Section("mongo-lonlat_data").Key("db").String()
			mconf_jw.Host = uc.iniFile.Section("mongo-lonlat_data").Key("host").String()
			mconf_jw.Port = uc.iniFile.Section("mongo-lonlat_data").Key("port").String()
			mconf_jw.UserName = uc.iniFile.Section("mongo-lonlat_data").Key("user").String()
			mconf_jw.UserPwd = uc.iniFile.Section("mongo-lonlat_data").Key("pwd").String()
			uc.debug_jw, _ = uc.iniFile.Section("mongo-lonlat_data").Key("debug").Int()
			uc.mp_jw = common.NewMgoPool(mconf_jw)

			//mysql 配置文件
			uc.mysql = orm.NewORM()

			uc.initData()
			uc.Doit(c)
		},
	}
}

func (this *UserCdTrace) Doit(c *cli.Context) {
	var (
		db        = this.iniFile.Section("mongo-data_source").Key("db").String()
		sess      = this.mp.Get()
		timestamp = common.GetDayTimestamp(-1) //0为今日
	)

	this.tagsByJwd = make(map[string]*TagInfo)

	iter := sess.DB(db).C(USERACTION_TABLE).Find(bson.M{"timestamp": timestamp}).Iter()
	// iter := sess.DB(db).C(USERACTION_TABLE).Find(bson.M{"timestamp": "1449417600"}).Limit(5).Iter()
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
func (this *UserCdTrace) getTagsJWDInfo(userInfo map[string]interface{}) {
	var (
		db   = this.iniFile.Section("mongo-lonlat_data").Key("db").String()
		sess = this.mp_jw.Get()
		ad   = userInfo["AD"].(string)
		err  error
	)
	defer sess.Close()

	var lonlatInfo map[string]interface{}
	err = sess.DB(db).C(JWD_TABLE).Find(bson.M{"ad": ad}).One(&lonlatInfo)
	if err != nil && err != mgo.ErrNotFound {
		log.Info(err)
		return
	}

	if len(lonlatInfo) > 0 {
		for _, tag := range userInfo["tag"].([]interface{}) { //获取每个ad内的tagid
			tagm := tag.(map[string]interface{})

			if tagm["tagmongo"].(string) == "1" { //如果是mongoid忽略
				continue
			}
			cid := tagm["tagId"].(string)
			if _, ok := this.taocat_list[cid]; !ok { //是否有这个标签，有时候标签格式会错误导致程序终止
				log.Info(cid)
				continue
			}
			cg := this.taocat_list[cid]
			if cg.Level != 3 { //从总标签的map判断是否是3级标签
				cid, err = cg.getLv3Id(this)
				if err != nil {
					continue
				}
			}

			lon := lonlatInfo["lon"].(string)
			lat := lonlatInfo["lat"].(string)

			key := cid + lon + lat
			if t, ok := this.tagsByJwd[key]; ok {
				t.num++
				continue
			}
			tmp_info := &TagInfo{
				tagid: cid,
				lon:   lon,
				lat:   lat,
				num:   1,
			}
			this.tagsByJwd[key] = tmp_info
		}
	}
}

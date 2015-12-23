package lonlat

import (
	"fmt"
	"github.com/ngaut/log"
	"github.com/qgweb/gopro/lib/encrypt"
	"github.com/qgweb/gopro/lib/grab"
	"github.com/qgweb/gopro/lib/orm"
	"io/ioutil"
	"runtime/debug"
	"time"
	// "time"

	"github.com/codegangsta/cli"
	"github.com/qgweb/gopro/9xutool/common"
	"gopkg.in/ini.v1"
	"gopkg.in/mgo.v2/bson"
)

func NewTagsNumberCli() cli.Command {
	return cli.Command{
		Name:  "daily_tags",
		Usage: "汇总昨日标签人数,操作tags_daily_report",
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
			uc.mp = common.NewMgoPool(mconf)
			//mysql 配置文件
			uc.mysql = orm.NewORM()

			uc.initData()
			uc.DataHandle(c)
		},
	}
}

func (this *UserCdTrace) DataHandle(c *cli.Context) {
	var (
		db    = this.iniFile.Section("mongo-data_source").Key("db").String()
		sess  = this.mp.Get()
		begin = common.GetDayTimestamp(-1)
		end   = common.GetDayTimestamp(0)
		err   error
		md5   encrypt.Md5
	)
	this.tags_num = make(map[string]int)
	this.uniqueUser = make(map[string]int)

	iter := sess.DB(db).C(USERACTION_TABLE).Find(bson.M{"timestamp": bson.M{"$gt": begin, "$lte": end}}).Iter() //昨天的数据
	i := 1
	for {
		var info map[string]interface{}
		if !iter.Next(&info) {
			break
		}

		ad := info["AD"].(string)
		ua := info["UA"].(string)

		for _, tag := range info["tag"].([]interface{}) { //获取每个ad内的tagid
			tagm := tag.(map[string]interface{})
			if tagm["tagmongo"].(string) == "1" { //如果是mongoid忽略
				continue
			}
			cid := tagm["tagId"].(string)
			m5 := md5.Encode(ad + ua + cid)
			if _, ok := this.uniqueUser[m5]; ok {
				continue
			}
			this.uniqueUser[m5] = 1
			if cg, ok := this.taocat_list[cid]; ok {
				if cg.Level != 3 { //从总标签的map判断是否是3级标签
					cid, err = cg.getLv3Id(this)
					if err != nil {
						continue
					}
				}
				this.tags_num[cid] = this.tags_num[cid] + 1
			}

		}
		fmt.Println("已处理", i, "条记录")
		i++
	}
	//排序
	fmt.Println("开始排序...")
	s_tags_num := grab.NewMapSorter(this.tags_num)
	s_tags_num.Sort()
	fmt.Println("排序完毕，开始插入数据")
	//入库
	if len(s_tags_num) > 0 {
		// if len(s_tags_num) > 50 { //取前50
		// 	s_tags_num = s_tags_num[0:50]
		// }
		this.getMysqlConnect() //连接mysql

		DayTimestamp := common.GetDayTimestamp(-1)
		NowTime := time.Now().Unix()
		for _, v := range s_tags_num {
			this.mysql.BSQL().Insert("tags_daily_report").Values("tag_id", "tag_text", "num", "create_time", "time")
			_, err := this.mysql.Insert(v.Key, this.taocat_list[v.Key].Name, v.Val, NowTime, DayTimestamp)
			if err != nil {
				log.Warn("插入失败 ", err)
			}
		}
	}
	log.Info("数据分析完毕!")
}

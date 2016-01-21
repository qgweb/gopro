package lonlat

import (
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/ngaut/log"
	"github.com/qgweb/gopro/9xutool/common"
	"github.com/qgweb/gopro/lib/orm"
	"gopkg.in/ini.v1"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"io/ioutil"
	"runtime/debug"
	"time"
)

type CarHouse struct {
	iniFile *ini.File
	mp      *common.MgoPool
	mysql   *orm.QGORM
	debug   int
}

type CarHouseData struct {
	lon      string //经度
	lat      string //纬度
	province string
	city     string
	district string
	time     string
	category int
	num      int
}

func NewCarHouseCli() cli.Command {
	return cli.Command{
		Name:  "daily_carhouse_by_cd",
		Usage: "根据ad坐标处理car和house,car_house_report_jw表",
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

			ch := &CarHouse{}
			ch.iniFile, _ = ini.Load(f)

			// mgo 配置文件
			mconf := &common.MgoConfig{}
			mconf.DBName = ch.iniFile.Section("mongo-industry").Key("db").String()
			mconf.Host = ch.iniFile.Section("mongo-industry").Key("host").String()
			mconf.Port = ch.iniFile.Section("mongo-industry").Key("port").String()
			mconf.UserName = ch.iniFile.Section("mongo-industry").Key("user").String()
			mconf.UserPwd = ch.iniFile.Section("mongo-industry").Key("pwd").String()
			ch.mp = common.NewMgoPool(mconf)
			//mysql 配置文件
			ch.mysql = orm.NewORM()

			ch.Do(c)
		},
	}
}

func (this *CarHouse) Do(c *cli.Context) {
	var (
		industry_db = this.iniFile.Section("mongo-industry").Key("db").String()
		lonlat_db   = this.iniFile.Section("mongo-lonlat_data").Key("db").String()
		collection  = getIndustryCollectionName()
		sess        = this.mp.Get()
		time        = common.GetDayTimestamp(-1)
	)
	defer sess.Close()

	iter := sess.DB(industry_db).C(collection).Find(bson.M{"timestamp": "1453305600"}).Limit(100).Iter()
	// iter := sess.DB(industry_db).C(collection).Find(bson.M{"timestamp": time}).Iter()
	var result map[string]interface{}
	var longlatData map[string]interface{}
	var CarHouseData_Map = make(map[string]*CarHouseData)
	var i = 1
	for {
		if !iter.Next(&result) {
			break
		}
		//car house表里的ad去匹配经纬度表的用户
		fmt.Println("正在匹配第", i, "条数据")
		i++
		ad := result["ad"].(string)

		err := sess.DB(lonlat_db).C("tbl_map").Find(bson.M{"ad": ad}).One(&longlatData)
		if err != nil && err != mgo.ErrNotFound { //如果有错误
			log.Fatal(err)
			continue
		}
		if err == mgo.ErrNotFound { //如果没有查询到
			continue
		}

		if _, ok := CarHouseData_Map[result["ad"].(string)]; ok {
			CarHouseData_Map[result["ad"].(string)].num = CarHouseData_Map[result["ad"].(string)].num + 1
			continue
		}
		var category int

		if result["category"].(string) == "car" {
			category = 10 //汽车10
		} else {
			category = 11
		}

		r := &CarHouseData{
			lon:      longlatData["lon"].(string),
			lat:      longlatData["lat"].(string),
			province: longlatData["province"].(string),
			city:     longlatData["city"].(string),
			district: longlatData["district"].(string),
			category: category,
			num:      1,
			time:     time,
		}
		CarHouseData_Map[result["ad"].(string)] = r
	}
	log.Info("数据处理完毕，开始入库...")

}

func getIndustryCollectionName() string {
	prefix := "industry_"
	datatime := time.Now().AddDate(0, 0, -1).Format("200601")
	dbname := prefix + datatime
	return dbname
}

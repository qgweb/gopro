package lonlat

import (
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/ngaut/log"
	"github.com/qgweb/gopro/9xutool/common"
	"gopkg.in/ini.v1"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"gopkg.in/olivere/elastic.v3"
	"io/ioutil"
	"runtime/debug"
	"strings"
	"time"
)

type CarHouse struct {
	iniFile     *ini.File
	industry_mp *common.MgoPool
	lonlat_mp   *common.MgoPool
	es          *elastic.Client
	debug       string
}

type CarHouseData struct {
	ad       string
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
			industry_mconf := &common.MgoConfig{}
			industry_mconf.DBName = ch.iniFile.Section("mongo-industry").Key("db").String()
			industry_mconf.Host = ch.iniFile.Section("mongo-industry").Key("host").String()
			industry_mconf.Port = ch.iniFile.Section("mongo-industry").Key("port").String()
			industry_mconf.UserName = ch.iniFile.Section("mongo-industry").Key("user").String()
			industry_mconf.UserPwd = ch.iniFile.Section("mongo-industry").Key("pwd").String()

			lonlat_mconf := &common.MgoConfig{}
			lonlat_mconf.DBName = ch.iniFile.Section("mongo-lonlat_data").Key("db").String()
			lonlat_mconf.Host = ch.iniFile.Section("mongo-lonlat_data").Key("host").String()
			lonlat_mconf.Port = ch.iniFile.Section("mongo-lonlat_data").Key("port").String()
			lonlat_mconf.UserName = ch.iniFile.Section("mongo-lonlat_data").Key("user").String()
			lonlat_mconf.UserPwd = ch.iniFile.Section("mongo-lonlat_data").Key("pwd").String()

			ch.industry_mp = common.NewMgoPool(industry_mconf)
			ch.lonlat_mp = common.NewMgoPool(lonlat_mconf)

			//debug
			ch.debug = ch.iniFile.Section("mongo-industry").Key("debug").String()
			//es
			esurl := ch.iniFile.Section("es").Key("host").String()
			ch.es, err = elastic.NewClient(elastic.SetURL(strings.Split(esurl, ",")...))
			if err != nil {
				log.Fatal(err)
			}
			ch.Do(c)
		},
	}
}

func (this *CarHouse) Do(c *cli.Context) {
	var (
		industry_db   = this.iniFile.Section("mongo-industry").Key("db").String()
		lonlat_db     = this.iniFile.Section("mongo-lonlat_data").Key("db").String()
		collection    = getIndustryCollectionName()
		industry_sess = this.industry_mp.Get()
		lonlat_sess   = this.lonlat_mp.Get()
		time          = common.GetDayTimestamp(-1)
	)
	defer industry_sess.Close()
	defer lonlat_sess.Close()

	// iter := industry_sess.DB(industry_db).C(collection).Find(bson.M{"timestamp": "1453305600"}).Limit(100).Iter()
	iter := industry_sess.DB(industry_db).C(collection).Find(bson.M{"timestamp": time}).Iter()
	var result map[string]interface{}
	var longlatData map[string]interface{}
	var CarHouseData_Map = make(map[string]*CarHouseData)
	var i = 1
	for {
		if !iter.Next(&result) {
			break
		}
		//car house表里的ad去匹配经纬度表的用户
		if this.debug == "1" {
			fmt.Println("正在匹配第", i, "条数据")
		}
		i++
		ad := result["ad"].(string)

		err := lonlat_sess.DB(lonlat_db).C("tbl_map").Find(bson.M{"ad": ad}).One(&longlatData)
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
			ad:       longlatData["ad"].(string),
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

	if len(CarHouseData_Map) > 0 {
		fmt.Println("数据处理完毕，开始入库...")
		bk := this.es.Bulk()
		for _, v := range CarHouseData_Map {
			bk.Add(elastic.NewBulkIndexRequest().Index("tags_car_house_report_jw").Type("map").Doc(map[string]interface{}{
				"ad":       v.ad,
				"geo":      v.lat + "," + v.lon,
				"tag_id":   v.category,
				"num":      v.num,
				"province": v.province,
				"city":     v.city,
				"district": v.district,
				"time":     v.time,
			}))
		}
		bk.Do()
	}
}

func getIndustryCollectionName() string {
	prefix := "industry_"
	datatime := time.Now().AddDate(0, 0, -1).Format("200601")
	dbname := prefix + datatime
	return dbname
}

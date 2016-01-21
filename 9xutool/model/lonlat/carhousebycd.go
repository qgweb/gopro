package lonlat

import (
	"github.com/ngaut/log"
	"github.com/qgweb/gopro/lib/orm"
	"gopkg.in/ini.v1"
	"runtime/debug"
)

type CarHouse struct {
	iniFile *ini.File
	mp      *common.MgoPool
	mysql   *orm.QGORM
	debug   int
}

type CarHouseData struct {
	Lot      string
	Lat      string
	Province string
	City     string
	District string
	Category int
	Num      int
	time     int
}

func NewCarHouseCli() cli.Command {
	return cli.Command{
		Name:  "daily_tags",
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
			mconf.DBName = ch.iniFile.Section("mongo-lonlat_data").Key("db").String()
			mconf.Host = ch.iniFile.Section("mongo-lonlat_data").Key("host").String()
			mconf.Port = ch.iniFile.Section("mongo-lonlat_data").Key("port").String()
			mconf.UserName = ch.iniFile.Section("mongo-lonlat_data").Key("user").String()
			mconf.UserPwd = ch.iniFile.Section("mongo-lonlat_data").Key("pwd").String()
			ch.mp = common.NewMgoPool(mconf)
			//mysql 配置文件
			ch.mysql = orm.NewORM()

			ch.Do(c)
		},
	}
}

func (ch *CarHouse) Do(c *cli.Context) {

}

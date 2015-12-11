package model

import (
	"fmt"
	"io/ioutil"
	"runtime/debug"
	"time"

	"github.com/codegangsta/cli"
	"github.com/ngaut/log"
	"github.com/qgweb/gopro/9xutool/common"
	"gopkg.in/ini.v1"
	"gopkg.in/mgo.v2/bson"
)

type (
	UserCdTrace struct {
		mp *common.MgoPool
		//		mysqlp  *common.MysqlPool	//待完成
		iniFile *ini.File
	}

	TaoCat struct { //数据模型
		Name  string
		Level int
		Cid   string
		Pid   string
	}

//	Tags_num
)

var (
	taocat_level_list map[int]map[string]*TaoCat //map[level]map[cid]TaoCat 用于取第三级标签信息
	taocat_list       map[string]*TaoCat         //标签分类总表 map[cid]TaoCat
	tags_num          = make(map[string]int)     //标签计数
)

func NewUserCdCli() cli.Command {
	return cli.Command{
		Name:  "get_tags_by_coordinate",
		Usage: "根据经纬度和ad汇总标签",
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

			ur := &UserCdTrace{}
			ur.iniFile, _ = ini.Load(f)

			// mgo 配置文件
			mconf := &common.MgoConfig{}
			mconf.DBName = ur.iniFile.Section("mongo").Key("db").String()
			mconf.Host = ur.iniFile.Section("mongo").Key("host").String()
			mconf.Port = ur.iniFile.Section("mongo").Key("port").String()
			mconf.UserName = ur.iniFile.Section("mongo").Key("user").String()
			mconf.UserPwd = ur.iniFile.Section("mongo").Key("pwd").String()
			ur.mp = common.NewMgoPool(mconf)

			ur.initData()
			//			ur.Do(c)
			ur.getTagsInfo(c)
			fmt.Println(tags_num)
		},
	}
}

func (this *UserCdTrace) initData() {
	db := this.iniFile.Section("mongo").Key("db").String()
	sess := this.mp.Get()
	defer sess.Close()

	var list []map[string]interface{}
	err := sess.DB(db).C("taocat").Find(bson.M{"type": "0"}).All(&list)
	if err != nil {
		log.Error(err)
	}

	taocat_level_list = make(map[int]map[string]*TaoCat)
	taocat_list = make(map[string]*TaoCat)
	if len(list) > 0 {
		for _, v := range list {
			//处理等级map
			category := &TaoCat{
				Name:  v["name"].(string),
				Level: v["level"].(int),
				Cid:   v["cid"].(string),
				Pid:   v["pid"].(string),
			}
			if _, ok := taocat_level_list[category.Level]; !ok {
				taocat_level_list[category.Level] = make(map[string]*TaoCat)
			}
			taocat_level_list[category.Level][category.Cid] = category
			taocat_list[category.Cid] = category
		}
	}

}

func (this *UserCdTrace) Do(c *cli.Context) {
}

//根据ad获取标签
func (this *UserCdTrace) getTagsInfo(c *cli.Context) {
	var (
		db       = this.iniFile.Section("mongo").Key("db").String()
		sess     = this.mp.Get()
		tagsInfo []map[string]interface{}
		//		dayTime  = getDay(0)
	)
	defer sess.Close()

	//	err := sess.DB(db).C("useraction").Find(bson.M{"AD": ad, "day": dayTime}).All(&tagsInfo)
	err := sess.DB(db).C("useraction").Find(bson.M{"AD": "YwdLb0cZUVlABmVXcAhgeg==", "day": "20151206"}).All(&tagsInfo)
	if err != nil {
		log.Error(err)
	}
	if len(tagsInfo) > 0 {
		for _, v := range tagsInfo { //可能会有多条数据，即多个ad
			for _, tag := range v["tag"].([]interface{}) { //获取每个ad内的tagid
				tagm := tag.(map[string]interface{})
				if tagm["tagmongo"].(string) == "1" { //如果是mongoid忽略
					continue
				}
				cid := tagm["tagId"].(string)
				cg := taocat_list[cid] //从总标签的map判断是否是3级标签
				if cg.Level != 3 {
					cid = cg.getLv3Id()
				}
				tags_num[cid] = tags_num[cid] + 1
			}
		}
	}
}

//获取相应的三级标签id
func (this *TaoCat) getLv3Id() string {
	if this.Level == 3 {
		return this.Cid
	}
	n := this.Level - 3
	tmp := *this
	for i := 0; i < n; i++ {
		tmp = *(taocat_list[tmp.Pid]) //获取父级
	}
	return tmp.Cid
}

//获取时间字符串
func getDay(day int) (tf string) {
	t := time.Now()
	if day != 0 {
		t = t.AddDate(0, 0, day)
	}
	tf = t.Format("20060102")
	return
}

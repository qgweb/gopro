package model

import (
	"fmt"
	// "gopkg.in/mgo.v2"
	"io/ioutil"
	"runtime/debug"

	"github.com/codegangsta/cli"
	"github.com/ngaut/log"
	"github.com/qgweb/gopro/9xutool/common"
	"gopkg.in/ini.v1"
	"gopkg.in/mgo.v2/bson"
)

type (
	UserCdTrace struct {
		mp      *common.MgoPool
		iniFile *ini.File
	}

	TaoCat struct { //数据模型
		Name  string
		Level int
		Cid   string
		Pid   string
	}
)

var taocat_list map[int]map[string]*TaoCat //map[level]map[cid]TaoCat 用于取第三级标签信息

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
			ur.Do(c)
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

	taocat_list = make(map[int]map[string]*TaoCat)
	// tmp_list = make
	if len(list) > 0 {
		for _, v := range list {
			t := TaoCat{}
			t.Name = v["name"].(string)
			t.Level = v["level"].(int)
			t.Cid = v["cid"].(string)
			t.Pid = v["pid"].(string)
			fmt.Println(t.Cid, t.Name)
			if _, ok := taocat_list[t.Level]; !ok {
				taocat_list[t.Level] = make(map[string]*TaoCat)
			}
			taocat_list[t.Level][t.Cid] = &t
		}
	}
	for k, v := range taocat_list {
		fmt.Println(k)
		for kk, vv := range v {
			fmt.Println(kk, vv.Name)
		}
	}

}

func (this *UserCdTrace) Do(c *cli.Context) {
}

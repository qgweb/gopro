package adcate

import (
	"fmt"
	"io/ioutil"
	"os"

	ossh "code.google.com/p/go.crypto/ssh"
	"github.com/codegangsta/cli"
	"github.com/ngaut/log"
	"github.com/qgweb/gopro/datatool/common/mongo"
	"github.com/qgweb/gopro/lib/grab"
	"github.com/qgweb/gopro/lib/ssh"
	"gopkg.in/ini.v1"
	"gopkg.in/mgo.v2/bson"
)

var (
	client  *ossh.Client
	mp      *mongo.MgoPool
	iniFile *ini.File
	err     error
	mgoConf *mongo.MgoConfig
)

func NewAdCate() cli.Command {
	return cli.Command{
		Name:  "adcate",
		Usage: "统计某天ad数量和分类各个数量",
		Action: func(c *cli.Context) {
			defer func() {
				if msg := recover(); msg != nil {
					log.Error(msg)
				}
			}()

			data, err := ioutil.ReadFile(c.String("conf"))
			if err != nil {
				log.Fatal(err)
				return
			}

			iniFile, err = ini.Load(data)
			if err != nil {
				log.Fatal(err)
				return
			}

			// ssh
			conf := ssh.Config{}
			conf.PrivaryKey = ssh.GetPrivateKey()
			conf.RemoteHost = iniFile.Section("ssh").Key("host").String()
			conf.RemotePort = 22
			conf.RemoteUser = "root"

			linker, err := mongo.GetSSHLinker(conf)
			if err != nil {
				log.Fatal(err)
				return
			}

			client = linker.GetClient()

			// mongo
			mgoConf = &mongo.MgoConfig{}
			mgoConf.Host = iniFile.Section("mongo").Key("host").String()
			mgoConf.Port = iniFile.Section("mongo").Key("port").String()
			mgoConf.UserName = iniFile.Section("mongo").Key("user").String()
			mgoConf.UserPwd = iniFile.Section("mongo").Key("pwd").String()
			mgoConf.DBName = iniFile.Section("mongo").Key("db").String()

			mp = &mongo.MgoPool{}

			Run()

		},
		Flags: []cli.Flag{
			cli.StringFlag{"conf", "./conf/ac.conf", "配置文件", ""},
		},
	}
}

func Run() {
	var (
		db    = iniFile.Section("mongo").Key("db").String()
		table = "useraction"
	)
	sess := mp.GetRemoteSession(mgoConf, client)
	defer sess.Close()

	list := make(map[string]int)
	list1 := make(map[string]map[string]int)
	list2 := make(map[string]int)
	var alist []map[string]interface{}
	//var blist []map[string]interface{}

	// 统计数量
	//	sess.DB(db).C(table).Find(bson.M{"day": "20151019"}).Select(bson.M{"_id": 0, "AD": 1}).All(&blist)
	//	for _, v := range blist {
	//		list2[v["AD"].(string)] = 1
	//	}

	err = sess.DB(db).C(table).Find(bson.M{"day": "20151019"}).Select(bson.M{"_id": 1, "tag": 1, "AD": 1}).All(&alist)
	log.Error(err)

	for _, v := range alist {
		for _, vv := range v["tag"].([]interface{}) {
			vvm := vv.(map[string]interface{})
			if vvm["tagId"].(string) == "" {
				continue
			}

			if _, ok := list1[vvm["tagId"].(string)]; !ok {
				list1[vvm["tagId"].(string)] = make(map[string]int)
			}

			list1[vvm["tagId"].(string)][v["AD"].(string)] = 1
		}
	}

	for k, v := range list1 {
		list[k] = len(v)
	}

	//16,1625,30,50011740,50006843,50006842,50010404
	var catMap = map[string]map[string]int{
		"16":       make(map[string]int),
		"1625":     make(map[string]int),
		"30":       make(map[string]int),
		"50011740": make(map[string]int),
		"50006843": make(map[string]int),
		"50006842": make(map[string]int),
		"50010404": make(map[string]int),
	}

	for k, _ := range list {
		var mm map[string]interface{}
		err = sess.DB(db).C("taocat").Find(bson.M{"cid": k}).Select(bson.M{"pid": 1, "_id": 0}).One(&mm)
		if err == nil {
			if _, ok := catMap[mm["pid"].(string)]; ok {
				for kk, _ := range list1[k] {
					catMap[mm["pid"].(string)][kk] = 1
				}
			}
		}
	}

	s := grab.NewMapSorter(list)
	s.Sort()
	log.Info(catMap)
	log.Error(s[0])
	log.Error(len(list2))

	for k, v := range catMap {
		f, _ := os.Create(k + ".txt")
		fmt.Println(k)
		for kk, _ := range v {
			f.WriteString(kk + "\n")
		}
		f.Close()
	}
}

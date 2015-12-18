package model

import (
	"github.com/codegangsta/cli"
	"github.com/ngaut/log"
	"github.com/qgweb/gopro/9xutool/common"
	"github.com/qgweb/gopro/lib/cache"
	"github.com/qgweb/gopro/lib/encrypt"
	"github.com/qgweb/gopro/lib/mongodb"
	"gopkg.in/ini.v1"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"io/ioutil"
	"strings"
	"time"
)

type URLTrace struct {
	mp      *mongodb.DialContext
	ldb     *cache.LevelDBCache
	iniFile *ini.File
}

func NewURLTraceCli() cli.Command {
	return cli.Command{
		Name:  "url_trace_merge",
		Usage: "生成域名1天的数据,供九旭精准投放",
		Action: func(c *cli.Context) {
			defer func() {
				if msg := recover(); msg != nil {
					log.Error(msg)
				}
			}()

			// 获取配置文件
			filePath := common.GetBasePath() + "/conf/ut.conf"
			f, err := ioutil.ReadFile(filePath)
			if err != nil {
				log.Fatal(err)
				return
			}

			ur := &URLTrace{}
			ur.iniFile, _ = ini.Load(f)

			// mgo 配置文件
			mconf := mongodb.MgoConfig{}
			mconf.DBName = ur.iniFile.Section("mongo").Key("db").String()
			mconf.Host = ur.iniFile.Section("mongo").Key("host").String()
			mconf.Port = ur.iniFile.Section("mongo").Key("port").String()
			mconf.UserName = ur.iniFile.Section("mongo").Key("user").String()
			mconf.UserPwd = ur.iniFile.Section("mongo").Key("pwd").String()
			ur.mp, err = mongodb.Dial(mongodb.GetLinkUrl(mconf), mongodb.GetCpuSessionNum())
			if err != nil {
				log.Fatal(err)
				return
			}
			ur.mp.Debug()
			// leveldb cache
			path := common.GetBasePath() + "/" + bson.NewObjectId().Hex()
			log.Info(path)
			if ur.ldb, err = cache.NewLevelDbCache(path); err != nil {
				log.Fatal(err)
				return
			}
			ur.Do(c)
			ur.mp.Close()
			ur.ldb.Close()
		},
	}
}

func (this *URLTrace) ReadData() {
	var (
		db     = this.iniFile.Section("mongo").Key("db").String()
		table  = "urltrack_" + time.Now().Format("200601")
		bgTime = common.GetHourTimestamp(-1)
		edTime = common.GetHourTimestamp(-2)
		query  = mongodb.MulQueryParam{}
	)

	query.DbName = db
	query.ColName = table
	query.Query = bson.M{"timestamp": bson.M{"$lte": bgTime, "$gte": edTime}}
	query.Fun = func(info map[string]interface{}) {
		ad := info["ad"].(string)
		ua := "ua"
		cids := info["cids"].([]interface{})
		if u, ok := info["ua"]; ok {
			ua = u.(string)
		}
		key := "adua_" + ad + "_" + ua

		for _, v := range cids {
			vm := v.(map[string]interface{})
			this.ldb.HSet(key, vm["id"].(string), "1")
			this.ldb.HSet("adua_keys", key, "1")
		}
	}
	this.mp.Query(query)
}

func (this *URLTrace) PutData() {
	var (
		db        = this.iniFile.Section("mongo").Key("db").String()
		table_put = "urltrack_put"
		sess      = this.mp.Ref()
	)
	defer this.mp.UnRef(sess)

	if keys, err := this.ldb.HGetAllKeys("adua_keys"); err == nil {
		list := make([]interface{}, 0, len(keys))
		for _, key := range keys {
			cids, err := this.ldb.HGetAllKeys(key)
			if err != nil {
				log.Error(err)
				continue
			}
			cidsmap := make([]bson.M, 0, len(cids))
			for _, cid := range cids {
				cidsmap = append(cidsmap, bson.M{"id": cid})
			}
			adua := strings.Split(strings.TrimPrefix(key, "adua_"), "_")
			list = append(list, bson.M{
				"ad":   adua[0],
				"ua":   adua[1],
				"adua": encrypt.DefaultMd5.Encode(adua[0] + adua[1]),
				"cids": cidsmap,
			})
		}
		log.Debug("共计:", len(list),"条")
		sess.DB(db).C(table_put).DropCollection()
		sess.DB(db).C(table_put).Create(&mgo.CollectionInfo{})
		sess.DB(db).C(table_put).EnsureIndexKey("cids.id")
		sess.DB(db).C(table_put).EnsureIndexKey("adua")
		this.mp.Insert(mongodb.MulQueryParam{db, table_put, nil, 0, nil}, list)
	}
}

func (this *URLTrace) Do(c *cli.Context) {
	this.ReadData()
	this.PutData()
}

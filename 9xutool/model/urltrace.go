package model

import (
	"github.com/codegangsta/cli"
	"github.com/ngaut/log"
	"github.com/qgweb/gopro/9xutool/common"
	"github.com/qgweb/gopro/lib/encrypt"
	"github.com/qgweb/gopro/lib/mongodb"
	"github.com/qgweb/gopro/lib/rediscache"
	"gopkg.in/ini.v1"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"io/ioutil"
	"strings"
	"sync"
	"time"
)

type URLTrace struct {
	mp        *mongodb.DialContext
	ldb       *rediscache.MemCache
	iniFile   *ini.File
	keyprefix string
	mux       sync.Mutex
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
			ur.keyprefix = mongodb.GetObjectId() + "_"
			// leveldb cache
			ur.initLevelDb()
			ur.Do(c)
			ur.mp.Close()
			ur.ldb.Clean(ur.keyprefix)
			ur.ldb.Clean(ur.keyprefix)
			ur.ldb.Close()
		},
	}
}

func (this *URLTrace) initLevelDb() {
	var err error
	config := rediscache.MemConfig{}
	config.Host = this.iniFile.Section("redis_cache").Key("host").String()
	config.Port = this.iniFile.Section("redis_cache").Key("port").String()

	if this.ldb, err = rediscache.New(config); err != nil {
		log.Fatal(err)
		return
	}
	this.ldb.SelectDb(this.iniFile.Section("redis_cache").Key("db").String())
}

func (this *URLTrace) SetData(key string, value string) {
	this.mux.Lock()
	this.ldb.HSet(this.keyprefix+key, value, "1")
	this.mux.Unlock()
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
	query.ChanSize = 1
	query.Query = bson.M{"timestamp": bson.M{"$lte": bgTime, "$gt": edTime}}
	query.Fun = func(info map[string]interface{}) {
		ad := info["ad"].(string)
		ua := "ua"
		cids := info["cids"].([]interface{})
		if u, ok := info["ua"]; ok {
			ua = u.(string)
		}
		key := ad + "_" + ua
		for _, v := range cids {
			vm := v.(map[string]interface{})
			this.SetData(key, vm["id"].(string))
		}
	}
	this.mp.Query(query)
	this.ldb.Flush()
}

func (this *URLTrace) PutData() {
	var (
		db        = this.iniFile.Section("mongo").Key("db").String()
		table_put = "urltrack_put"
		sess      = this.mp.Ref()
	)
	defer this.mp.UnRef(sess)

	keys := this.ldb.Keys(this.keyprefix + "*")
	list := make([]interface{}, 0, len(keys))
	log.Debug(len(keys))
	for _, key := range keys {
		cids := this.ldb.HGetAllKeys(key)
		cidsmap := make([]bson.M, 0, len(cids))
		for _, cid := range cids {
			cidsmap = append(cidsmap, bson.M{"id": cid})
		}
		adua := strings.Split(strings.TrimPrefix(key, this.keyprefix), "_")
		list = append(list, bson.M{
			"ad":   adua[0],
			"ua":   adua[1],
			"adua": encrypt.DefaultMd5.Encode(adua[0] + adua[1]),
			"cids": cidsmap,
		})
		if len(list)%10000 == 0 {
			log.Error(len(list))
		}
	}
	log.Debug("共计:", len(list), "条")
	sess.DB(db).C(table_put).DropCollection()
	sess.DB(db).C(table_put).Create(&mgo.CollectionInfo{})
	sess.DB(db).C(table_put).EnsureIndexKey("cids.id")
	sess.DB(db).C(table_put).EnsureIndexKey("adua")
	this.mp.Insert(mongodb.MulQueryParam{db, table_put, nil, 0, nil, 1}, list)
}

func (this *URLTrace) Do(c *cli.Context) {
	this.ReadData()
	this.PutData()
}

package model

import (
	"github.com/codegangsta/cli"
	"github.com/garyburd/redigo/redis"
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

type UserTrace struct {
	mp      *mongodb.DialContext
	mcp     *mongodb.DialContext
	mtcp    *mongodb.DialContext
	ldb     *rediscache.MemCache
	rc      redis.Conn
	iniFile *ini.File
	prefix  string
	mux     sync.Mutex
}

func NewUserTraceCli() cli.Command {
	return cli.Command{
		Name:  "user_trace_merge",
		Usage: "生成用户最近3天浏览轨迹,供九旭精准投放",
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

			ur := &UserTrace{}
			ur.iniFile, _ = ini.Load(f)

			// mgo 配置文件
			if ur.mp, err = ur.GetMongoObj("mongo"); err != nil {
				log.Fatal(err)
				return
			}
			if ur.mcp, err = ur.GetMongoObj("mongo-cookie"); err != nil {
				log.Fatal(err)
				return
			}
			if ur.mtcp, err = ur.GetMongoObj("mongo-tj-cookie"); err != nil {
				log.Fatal(err)
				return
			}
			// leveldb cache
			ur.prefix = bson.NewObjectId().Hex() + "_"
			ur.initLevelDb()
			ur.Do(c)
			ur.mp.Close()
			ur.mcp.Close()
			ur.mtcp.Close()
			ur.ldb.Clean(ur.prefix)
			ur.ldb.Clean(ur.prefix)
			ur.ldb.Close()
		},
	}
}

func (this *UserTrace) initLevelDb() {
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

func (this *UserTrace) GetMongoObj(section string) (*mongodb.DialContext, error) {
	mconf2 := mongodb.MgoConfig{}
	mconf2.DBName = this.iniFile.Section(section).Key("db").String()
	mconf2.Host = this.iniFile.Section(section).Key("host").String()
	mconf2.Port = this.iniFile.Section(section).Key("port").String()
	mconf2.UserName = this.iniFile.Section(section).Key("user").String()
	mconf2.UserPwd = this.iniFile.Section(section).Key("pwd").String()
	return mongodb.Dial(mongodb.GetLinkUrl(mconf2), mongodb.GetCpuSessionNum())
}

func (this *UserTrace) setInfo(key string, value string) {
	this.mux.Lock()
	key = this.prefix + key
	this.ldb.HSet(key, value, "1")
	this.mux.Unlock()
}

// cookie白名单数据
// 格莱美:55e5661525d0a2091a567a70
// 7k7k : 55e5661525d0a2091a567a64
func (this *UserTrace) WhiteCookie() {
	var (
		tag         = "55e5661525d0a2091a567a70"
		param       = mongodb.MulQueryParam{}
		sess        = this.mtcp.Ref()
		sess_cookie = this.mcp.Ref()
	)
	defer this.mtcp.UnRef(sess)
	defer this.mcp.UnRef(sess_cookie)
	log.Info("WhiteCookie begin")
	param.DbName = "xu_tj"
	param.ColName = "cookie"
	param.Query = bson.M{}
	param.Fun = func(info map[string]interface{}) {
		var adua map[string]interface{}
		ck := info["ck"].(string)
		if !bson.IsObjectIdHex(ck) {
			return
		}
		bt := time.Now()
		err := sess_cookie.DB("user_cookie").C("dt_user").Find(bson.M{"_id": bson.ObjectIdHex(ck)}).
			Select(bson.M{"cox": 1, "ua": 1}).One(&adua)
		log.Warn(time.Now().Sub(bt).Seconds())
		if err != nil {
			log.Error(err)
			return
		}

		ad := adua["cox"].(string)
		ua := encrypt.DefaultBase64.Encode(adua["ua"].(string))
		referer := info["referer"].(string)
		if ad == "" {
			return
		}

		if strings.Contains(referer, "7k7k.com") || strings.Contains(referer, "95k.com") {
			tag = "55e5661525d0a2091a567a64"
		}

		this.setInfo(ad+"_"+ua, tag)
	}
	this.mtcp.Query(param)
	this.ldb.Flush()
	log.Info("WhiteCookie ok")
}

func (this *UserTrace) ReadData(query bson.M) {
	var (
		db    = this.iniFile.Section("mongo").Key("db").String()
		table = "useraction"
		param = mongodb.MulQueryParam{}
	)
	log.Debug(query)

	param.DbName = db
	param.ColName = table
	param.Query = query
	param.Fun = func(info map[string]interface{}) {
		ua := "ua"
		ad := info["AD"].(string)
		if u, ok := info["UA"]; ok {
			ua = u.(string)
		}
		for _, v := range info["tag"].([]interface{}) {
			if tags, ok := v.(map[string]interface{}); ok {
				this.setInfo(ad+"_"+ua, tags["tagId"].(string))
			}
		}
	}
	this.mp.Query(param)
	this.ldb.Flush()
}

// 获取大分类
func (this *UserTrace) getBigCat() map[string]string {
	var (
		db    = this.iniFile.Section("mongo").Key("db").String()
		table = "taocat"
		sess  = this.mp.Ref()
	)

	defer this.mp.UnRef(sess)

	var info []map[string]interface{}
	var list = make(map[string]string)
	err := sess.DB(db).C(table).Find(bson.M{"type": "0"}).Select(bson.M{"bid": 1, "cid": 1}).All(&info)
	if err == nil {
		for _, v := range info {
			list[v["cid"].(string)] = v["bid"].(string)
		}
		return list
	}
	return nil
}

func (this *UserTrace) saveData() {
	var (
		sess          = this.mp.Ref()
		db            = this.iniFile.Section("mongo").Key("db").String()
		table_put     = "useraction_put"
		table_put_big = "useraction_put_big"
		taoCategory   map[string]string
	)

	defer this.mp.UnRef(sess)

	//初始化淘宝分类
	taoCategory = this.getBigCat()
	keys := this.ldb.Keys(this.prefix + "*")
	list := make([]interface{}, 0, len(keys))
	list_put := make([]interface{}, 0, len(keys))
	for _, key := range keys {
		tags := this.ldb.HGetAllKeys(key)
		tagsmap := make([]bson.M, 0, len(tags))
		tagsmap_put := make([]bson.M, 0, len(tags))
		for _, tag := range tags {
			taginfo := bson.M{"tagId": tag, "tagmongo": "0"}
			taginfo_put := bson.M{"tagId": tag, "tagmongo": "0"}
			if bson.IsObjectIdHex(tag) {
				taginfo["tagmongo"] = "1"
				taginfo_put["tagmongo"] = "1"
			}
			if bt, ok := taoCategory[tag]; ok {
				taginfo_put["tagId"] = bt
			}
			tagsmap = append(tagsmap, taginfo)
			tagsmap_put = append(tagsmap_put, taginfo_put)
		}

		adua := strings.Split(strings.TrimPrefix(key, this.prefix), "_")
		list = append(list, bson.M{
			"AD":   adua[0],
			"UA":   adua[1],
			"adua": encrypt.DefaultMd5.Encode(adua[0] + adua[1]),
			"tag":  tagsmap,
		})
		list_put = append(list_put, bson.M{
			"AD":   adua[0],
			"UA":   adua[1],
			"adua": encrypt.DefaultMd5.Encode(adua[0] + adua[1]),
			"tag":  tagsmap_put,
		})
	}

	log.Debug("共计:", len(list), "条")
	//加索引
	sess.DB(db).C(table_put).DropCollection()
	sess.DB(db).C(table_put_big).DropCollection()
	sess.DB(db).C(table_put).Create(&mgo.CollectionInfo{})
	sess.DB(db).C(table_put_big).Create(&mgo.CollectionInfo{})
	sess.DB(db).C(table_put).EnsureIndexKey("tag.tagId")
	sess.DB(db).C(table_put).EnsureIndexKey("adua")
	sess.DB(db).C(table_put).EnsureIndexKey("AD")
	sess.DB(db).C(table_put_big).EnsureIndexKey("tag.tagId")
	sess.DB(db).C(table_put_big).EnsureIndexKey("adua")
	sess.DB(db).C(table_put_big).EnsureIndexKey("AD")
	this.mp.Insert(mongodb.MulQueryParam{db, table_put, nil, 0, nil, 1}, list)
	this.mp.Insert(mongodb.MulQueryParam{db, table_put_big, nil, 0, nil, 1}, list_put)
}

func (this *UserTrace) Do(c *cli.Context) {
	var (
		eghour = common.GetHourTimestamp(-1)
		bghour = common.GetHourTimestamp(-2)
		egdate = common.GetHourTimestamp(-1)
		bgdate = common.GetHourTimestamp(-73)
	)
	this.mp.Debug()
	this.ReadData(bson.M{"timestamp": bson.M{"$gte": bghour, "$lte": eghour}, "domainId": bson.M{"$ne": "0"}})
	this.ReadData(bson.M{"timestamp": bson.M{"$gte": bgdate, "$lte": egdate}, "domainId": "0"})
	//this.WhiteCookie()
	this.saveData()
}

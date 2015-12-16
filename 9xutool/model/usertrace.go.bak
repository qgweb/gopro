package model

import (
	"fmt"
	"io/ioutil"
	"math"
	"strings"
	"sync"
	"time"

	"github.com/qgweb/gopro/lib/encrypt"

	"github.com/codegangsta/cli"
	"github.com/garyburd/redigo/redis"
	"github.com/ngaut/log"
	"github.com/qgweb/gopro/9xutool/common"
	"github.com/qgweb/gopro/lib/convert"
	"gopkg.in/ini.v1"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type UserTrace struct {
	mp      *common.MgoPool
	mcp     *common.MgoPool
	mtcp    *common.MgoPool
	rc      redis.Conn
	iniFile *ini.File
	prefix  string
}

type WaitGroup struct {
	sync.WaitGroup
}

func (this *WaitGroup) Run(fun func(...interface{}), param ...interface{}) {
	this.Add(1)
	go func() {
		fun(param...)
		this.Done()
	}()
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
			mconf := &common.MgoConfig{}
			mconf.DBName = ur.iniFile.Section("mongo").Key("db").String()
			mconf.Host = ur.iniFile.Section("mongo").Key("host").String()
			mconf.Port = ur.iniFile.Section("mongo").Key("port").String()
			mconf.UserName = ur.iniFile.Section("mongo").Key("user").String()
			mconf.UserPwd = ur.iniFile.Section("mongo").Key("pwd").String()
			ur.mp = common.NewMgoPool(mconf)

			mconf1 := &common.MgoConfig{}
			mconf1.DBName = ur.iniFile.Section("mongo-cookie").Key("db").String()
			mconf1.Host = ur.iniFile.Section("mongo-cookie").Key("host").String()
			mconf1.Port = ur.iniFile.Section("mongo-cookie").Key("port").String()
			mconf1.UserName = ur.iniFile.Section("mongo-cookie").Key("user").String()
			mconf1.UserPwd = ur.iniFile.Section("mongo-cookie").Key("pwd").String()
			ur.mcp = common.NewMgoPool(mconf1)

			mconf2 := &common.MgoConfig{}
			mconf2.DBName = ur.iniFile.Section("mongo-tj-cookie").Key("db").String()
			mconf2.Host = ur.iniFile.Section("mongo-tj-cookie").Key("host").String()
			mconf2.Port = ur.iniFile.Section("mongo-tj-cookie").Key("port").String()
			mconf2.UserName = ur.iniFile.Section("mongo-tj-cookie").Key("user").String()
			mconf2.UserPwd = ur.iniFile.Section("mongo-tj-cookie").Key("pwd").String()
			ur.mtcp = common.NewMgoPool(mconf2)

			// redis配置
			ruser := ur.iniFile.Section("redis").Key("host").String()
			rport := ur.iniFile.Section("redis").Key("port").String()
			rdb := ur.iniFile.Section("redis").Key("db").String()
			ur.rc, err = redis.Dial("tcp4", fmt.Sprintf("%s:%s", ruser, rport))
			if err != nil {
				log.Fatal(err)
				return
			}
			ur.rc.Do("SELECT", rdb)

			ur.prefix = bson.NewObjectId().Hex() + "_"
			ur.Do(c)
		},
	}
}

func (this *UserTrace) setInfo(ad string, tagid string) {
	this.rc.Do("HSET", this.prefix+ad, tagid, 1)
}

func (this *UserTrace) getall(key string) map[string]string {
	list, err := redis.StringMap(this.rc.Do("HGETALL", key))
	if err != nil {
		return nil
	}
	return list
}

func (this *UserTrace) emptyKeys() {
	keys, err := redis.Strings(this.rc.Do("keys", this.prefix+"*"))
	if err == nil {
		for _, v := range keys {
			this.rc.Do("DEL", v)
		}
	}
}

func (this *UserTrace) saveData() {
	var (
		size          = 10000
		page          = 1
		pageTotal     = 0
		pageCount     = 0
		sess          = this.mp.Get()
		db            = this.iniFile.Section("mongo").Key("db").String()
		table_put     = "useraction_put"
		table_put_big = "useraction_put_big"
		taoCategory   map[string]string
	)

	defer sess.Close()

	sess.DB(db).C(table_put).DropCollection()
	sess.DB(db).C(table_put_big).DropCollection()

	//加索引
	sess.DB(db).C(table_put).Create(&mgo.CollectionInfo{})
	sess.DB(db).C(table_put_big).Create(&mgo.CollectionInfo{})
	sess.DB(db).C(table_put).EnsureIndexKey("tag.tagId")
	sess.DB(db).C(table_put).EnsureIndexKey("adua")
	sess.DB(db).C(table_put_big).EnsureIndexKey("tag.tagId")
	sess.DB(db).C(table_put_big).EnsureIndexKey("adua")
	//初始化淘宝分类
	taoCategory = this.getBigCat()

	keys, err := redis.Strings(this.rc.Do("KEYS", this.prefix+"*"))
	if err != nil {
		return
	}

	pageCount = len(keys)

	pageTotal = int(math.Ceil(float64(pageCount) / float64(size)))
	for page = 1; page <= pageTotal; page++ {
		bg := (page - 1) * size
		eg := bg + size
		if page == pageTotal {
			eg = pageCount
		}

		var list []interface{} = make([]interface{}, 0, eg-bg)
		var list_put []interface{} = make([]interface{}, 0, eg-bg)
		for _, v := range keys[bg:eg] {
			tags := this.getall(v)
			info := make(bson.M)
			info_put := make(bson.M)
			v = strings.Replace(v, this.prefix, "", -1)
			adua := strings.Split(v, "_")

			info["AD"] = adua[0]
			info_put["AD"] = adua[0]
			info["UA"] = adua[1]
			info_put["UA"] = adua[1]
			info["adua"] = encrypt.DefaultMd5.Encode(adua[0] + adua[1])
			info_put["adua"] = encrypt.DefaultMd5.Encode(adua[0] + adua[1])
			info["tag"] = make([]bson.M, 0, len(tags))
			info_put["tag"] = make([]bson.M, 0, len(tags))

			for kk, _ := range tags {
				tagInfo := make(bson.M)
				tagInfoPut := make(bson.M)

				tagInfo["tagId"] = kk
				tagInfo["tagmongo"] = "0"
				tagInfoPut["tagId"] = kk
				tagInfoPut["tagmongo"] = "0"

				if bson.IsObjectIdHex(kk) {
					tagInfo["tagmongo"] = "1"
					tagInfoPut["tagmongo"] = "1"
				} else {
					if bt, ok := taoCategory[kk]; ok {
						tagInfoPut["tagId"] = bt
					}
				}

				info["tag"] = append(info["tag"].([]bson.M), tagInfo)
				info_put["tag"] = append(info_put["tag"].([]bson.M), tagInfoPut)
			}

			list = append(list, info)
			list_put = append(list_put, info_put)
		}

		log.Info(len(list))
		//插入mongo
		sess.DB(db).C(table_put).Insert(list...)
		sess.DB(db).C(table_put_big).Insert(list_put...)
	}
}

func (this *UserTrace) ReadData(query bson.M) {
	var (
		db    = this.iniFile.Section("mongo").Key("db").String()
		sess  = this.mp.Get()
		table = "useraction"
	)
	log.Debug(query)
	iter := sess.DB(db).C(table).Find(query).Iter()
	var data map[string]interface{}
	for {
		if !iter.Next(&data) {
			break
		}

		for _, v := range data["tag"].([]interface{}) {
			if tags, ok := v.(map[string]interface{}); ok {
				ua := "ua"
				ad := data["AD"].(string)
				if u, ok := data["UA"]; ok {
					ua = u.(string)
				}
				this.setInfo(ad+"_"+ua, tags["tagId"].(string))
			}
		}
	}
}

// 获取大分类
func (this *UserTrace) getBigCat() map[string]string {
	var (
		db    = this.iniFile.Section("mongo").Key("db").String()
		table = "taocat"
		sess  = this.mp.Get()
	)

	defer sess.Close()

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

// cookie白名单数据
// 格莱美:55e5661525d0a2091a567a70
func (this *UserTrace) WhiteCookie() {
	var (
		sess    = this.mcp.Get()
		sess_tj = this.mtcp.Get()

		//btime    = time.Now().Add(-time.Hour * 24).Format("2006-01-02")
		//bdate, _ = time.ParseInLocation("2006-01-02", btime, time.Local)
		tag = "55e5661525d0a2091a567a70"
		num = 0
	)

	defer sess.Close()
	defer sess_tj.Close()

	iter := sess_tj.DB("xu_tj").C("cookie").Find(nil).Select(bson.M{"ck": 1, "_id": 0}).Iter()

	for {
		var info map[string]interface{}
		var adua map[string]interface{}
		var err error

		if !iter.Next(&info) {
			log.Error(111)
			break
		}
		ck := info["ck"].(string)
		if !bson.IsObjectIdHex(ck) {
			continue
		}

		err = sess.DB("user_cookie").C("dt_user").
			Find(bson.M{"_id": bson.ObjectIdHex(ck)}).
			Select(bson.M{"cox": 1, "ua": 1}).One(&adua)
		if err != nil {
			log.Error(err)
			continue
		}
		ad := adua["cox"].(string)
		ua := encrypt.DefaultBase64.Encode(adua["ua"].(string))

		if ad == "" {
			continue
		}
		this.setInfo(ad+"_"+ua, tag)
		num = num + 1
		log.Info(num)
	}
}

func (this *UserTrace) Do(c *cli.Context) {
	var (
		now    = time.Now()
		now1   = now.Add(-time.Second * time.Duration(now.Second())).Add(-time.Minute * time.Duration(now.Minute()))
		eghour = convert.ToString(now1.Add(-time.Hour).Unix())
		bghour = convert.ToString(now1.Add(-time.Duration(time.Hour * 2)).Unix())
		egdate = convert.ToString(now1.Add(-time.Hour).Unix())
		bgdate = convert.ToString(now1.Add(-time.Duration(time.Hour * 73)).Unix())
	)

	this.WhiteCookie()
	this.ReadData(bson.M{"domainId": bson.M{"$ne": "0"}, "timestamp": bson.M{"$gte": bghour, "$lte": eghour}})
	this.ReadData(bson.M{"domainId": "0", "timestamp": bson.M{"$gte": bgdate, "$lte": egdate}})

	this.saveData()
	this.emptyKeys()
}

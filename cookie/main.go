package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"strings"
	"time"

	"sync"

	"goclass/encrypt"
	"goclass/grab"
	"goclass/orm"

	"github.com/garyburd/redigo/redis"
	"gopkg.in/ini.v1"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var (
	IniFile    *ini.File
	err        error
	conf       = flag.String("conf", "conf.ini", "配置文件")
	mux        sync.Mutex
	mdbsession *mgo.Session
	msqldb     *orm.QGORM
	ptags      []string //投放标签集合
	rpool      *redis.Pool
)

func init() {
	flag.Parse()
	data, err := ioutil.ReadFile(*conf)
	if err != nil {
		log.Fatalln("读取配置文件失败,错误信息为:", err)
	}

	IniFile, err = ini.Load(data)
	if err != nil {
		log.Fatalln("加载配置文件内容失败,错误信息为:", err)
	}

	//初始化mysql链接
	GetMysqlConn()
	// 读取投放标签集合
	initPutTags()
	//获取redis链接池
	rpool = GetRedisPool()
}

func main() {
	for {
		c := time.After(time.Hour)
		t := <-c
		ReadTags()
		fmt.Println("执行时间:", t.Format("2006-01-02 15:04:05"))
	}
}

// 读取cookie对应标签信息
func ReadTags() {
	sess := GetSession()
	defer sess.Close()

	var (
		modb      = IniFile.Section("mongo").Key("db").String()
		prefix    = IniFile.Section("default").Key("prefix").String()
		tableName = prefix + "_cookie_tags_put"
		limit     = 1000
	)
	total, err := sess.DB(modb).C(tableName).Find(bson.M{}).Count()
	if err != nil {
		log.Println("获取mongo-标签总量失败,错误信息为:", err)
		return
	}

	for i := 1; i <= int(math.Ceil(float64(total)/float64(limit))); i++ {
		var list []map[string]interface{}
		err = sess.DB(modb).C(tableName).Find(bson.M{}).Skip((i - 1) * limit).Limit(limit).All(&list)
		if err != nil {
			log.Println("获取mongo-标签数据失败,错误信息为:", err)
			continue
		}
		for _, v := range list {
			cookie := v["cookie"].(string)
			tags := v["cids"].([]interface{})
			tagsAry := make([]string, 0, 10)
			for _, b := range tags {
				if c, ok := b.(map[string]interface{}); ok {
					if grab.In_Array(ptags, c["tagid"].(string)) {
						tagsAry = append(tagsAry, c["tagid"].(string))
					}
				}
			}
			if len(tagsAry) > 0 {
				advertIds := ReadAdvertByTag(tagsAry)
				if len(advertIds) > 0 {
					pushData(cookie, advertIds)
				}
			}

		}
	}
}

//读取标签对应的广告
func ReadAdvertByTag(tagsAry []string) []string {
	var (
		strategyid = IniFile.Section("default").Key("strategyid").String()
	)
	//SELECT * FROM `nxu_advert` WHERE `strategy_id` =8221 AND isdel=0 AND (`begin_time` <= {$now} <=`end_time`) AND (`title` LIKE '%{$keyword}%') limit 15
	where := fmt.Sprintf("`strategy_id` =%s AND isdel=0 AND (`begin_time` <= %d <=`end_time`) AND (`tag_id` in (%s))",
		strategyid, time.Now().Unix(), strings.Join(tagsAry, ","))
	msqldb.BSQL().Select("id").From("nxu_advert").Where(where).Limit(0, 15)
	list, err := msqldb.Query()
	if err != nil {
		fmt.Println("读取mysql数据失败,错误信息为:", err, ",sql为:", msqldb.LastSql())
	}

	advertIds := make([]string, 0, 15)
	for _, v := range list {
		advertIds = append(advertIds, v["id"])
	}

	return advertIds
}

//推送数据到redis
func pushData(cookie string, advertIds []string) {
	conn := rpool.Get()
	defer conn.Close()
	var (
		db  = IniFile.Section("redis").Key("db").String()
		key = "COOKIE_TAGS_" + cookie
	)

	conn.Do("SELECT", db)
	_, err = conn.Do("SET", key, strings.Join(advertIds, ","))
	if err != nil {
		log.Println("推送redis失败,", err)
	}
	conn.Do("EXPIRE", key, 24*3600)
}

// 读取投放标签集合
func initPutTags() {
	tags := IniFile.Section("default").Key("tags").String()
	ptags = strings.Split(tags, ",")
}

//获取mysql链接
func GetMysqlConn() {
	var (
		mouser = IniFile.Section("mysql").Key("user").String()
		mopwd  = IniFile.Section("mysql").Key("pwd").String()
		mohost = IniFile.Section("mysql").Key("host").String()
		moport = IniFile.Section("mysql").Key("port").String()
		modb   = IniFile.Section("mysql").Key("db").String()
	)

	mopwd = encrypt.GetEnDecoder(encrypt.TYPE_BASE64).Decode(mopwd)

	msqldb = orm.NewORM()
	err := msqldb.Open(fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8", mouser, mopwd, mohost, moport, modb))
	if err != nil {
		log.Fatalln("初始化mysql链接失败,错误信息:", err)
	}
}

//redis链接
func GetRedisPool() *redis.Pool {
	var (
		host = IniFile.Section("redis").Key("host").String()
		port = IniFile.Section("redis").Key("port").String()
	)

	return &redis.Pool{
		MaxIdle:   20,
		MaxActive: 20, // max number of connections
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", host+":"+port)
			if err != nil {
				log.Fatalln(err.Error())
			}
			return c, err
		},
	}
}

//获取mongo数据库链接
func GetSession() *mgo.Session {
	mux.Lock()
	defer mux.Unlock()

	var (
		mouser = IniFile.Section("mongo").Key("user").String()
		mopwd  = IniFile.Section("mongo").Key("pwd").String()
		mohost = IniFile.Section("mongo").Key("host").String()
		moport = IniFile.Section("mongo").Key("port").String()
		modb   = IniFile.Section("mongo").Key("db").String()
	)

	if mdbsession == nil {
		var err error
		mdbsession, err = mgo.Dial(fmt.Sprintf("%s:%s@%s:%s/%s", mouser, mopwd, mohost, moport, modb))
		if err != nil {
			panic(err)
		}
	}
	//高并发下会关闭连接,ping下会恢复
	mdbsession.Ping()
	return mdbsession.Copy()
}

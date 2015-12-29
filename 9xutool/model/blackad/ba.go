package blackad

import (
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/garyburd/redigo/redis"
	"github.com/ngaut/log"
	"github.com/qgweb/gopro/9xutool/common"
	"github.com/qgweb/gopro/lib/convert"
	"github.com/qgweb/gopro/lib/encrypt"
	"github.com/qgweb/gopro/lib/mongodb"
	"github.com/qgweb/gopro/lib/rediscache"
	"gopkg.in/ini.v1"
	"gopkg.in/mgo.v2/bson"
	"io/ioutil"
	"os"
	"runtime/debug"
	"strings"
	"sync"
	"time"
)

// 广告黑名单生成
// 找出点击比较低的广告
type BlackMenu struct {
	iniFile         *ini.File
	mp              *mongodb.DialContext
	rc_put          redis.Conn
	blackprefix     string         // 黑名单
	provinceAdverts map[string]int // 浙江广告集合
	proprefix       string         // 浙江省对应的广告前缀
	blackFilePath   string         // 黑名单存放目录
	blackFileName   string         // 黑名单文件
	ldb             *rediscache.MemCache
	mux             sync.Mutex
	keyprefix       string
}

// 获取monggo对象
func getMongoObj(inifile *ini.File) (*mongodb.DialContext, error) {
	mconf := mongodb.MgoConfig{}
	mconf.DBName = inifile.Section("mongo").Key("db").String()
	mconf.Host = inifile.Section("mongo").Key("host").String()
	mconf.Port = inifile.Section("mongo").Key("port").String()
	mconf.UserName = inifile.Section("mongo").Key("user").String()
	mconf.UserPwd = inifile.Section("mongo").Key("pwd").String()
	return mongodb.Dial(mongodb.GetLinkUrl(mconf), mongodb.GetCpuSessionNum())
}

// 获取redis对象
func getRedisObj(section string, inifile *ini.File) redis.Conn {
	ruser := inifile.Section(section).Key("host").String()
	rport := inifile.Section(section).Key("port").String()
	rdb := inifile.Section(section).Key("db").String()
	rauth := inifile.Section(section).Key("auth").String()
	rc, err := redis.Dial("tcp4", fmt.Sprintf("%s:%s", ruser, rport))
	if err != nil {
		log.Fatal(err)
		return nil
	}
	if strings.TrimSpace(rauth) != "" {
		rc.Do("AUTH", "qaz#qiguan#wsx") //电信密钥
	}
	rc.Do("SELECT", rdb)
	return rc
}

func getConfig(name string) *ini.File {
	// 获取配置文件
	filePath := common.GetBasePath() + "/conf/" + name
	f, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Fatal(err)
		return nil
	}
	iniFile, err := ini.Load(f)
	if err != nil {
		log.Fatal(err)
		return nil
	}

	return iniFile
}

func NewBlackMenuCli() cli.Command {
	return cli.Command{
		Name:  "create_blackmenu",
		Usage: "生成黑名单广告的数据",
		Action: func(c *cli.Context) {
			defer func() {
				if msg := recover(); msg != nil {
					log.Error(msg)
					debug.PrintStack()
				}
			}()
			var err error
			ur := &BlackMenu{}
			ur.iniFile = getConfig("black.conf")
			if ur.mp, err = getMongoObj(ur.iniFile); err != nil {
				log.Fatal(err)
				return
			}
			ur.rc_put = getRedisObj("redis_put", ur.iniFile)
			ur.blackprefix = ur.iniFile.Section("default").Key("black_prefix").String()
			ur.proprefix = ur.iniFile.Section("default").Key("province_prefix").String()
			ur.blackFileName = ur.iniFile.Section("default").Key("black_file_put_name").String()
			ur.blackFilePath = ur.iniFile.Section("default").Key("balck_file_path").String()
			ur.ldb = ur.initLevelDb()
			ur.keyprefix = mongodb.GetObjectId() + "_"
			ur.Do(c)
			ur.ldb.Clean(ur.keyprefix)
			ur.ldb.Clean(ur.keyprefix)
			ur.rc_put.Close()
			ur.ldb.Close()
		},
	}
}

func (this *BlackMenu) initLevelDb() *rediscache.MemCache {
	config := rediscache.MemConfig{}
	config.Host = this.iniFile.Section("redis_cache").Key("host").String()
	config.Port = this.iniFile.Section("redis_cache").Key("port").String()
	ldb, err := rediscache.New(config)
	if err != nil {
		log.Fatal(err)
		return nil
	}
	ldb.SelectDb(this.iniFile.Section("redis_cache").Key("db").String())
	return ldb
}

func (this *BlackMenu) setDataToDb(key string, value string) {
	this.mux.Lock()
	this.ldb.Set(this.keyprefix+key, value)
	this.mux.Unlock()
}

//把黑名单保存到文件中
func (this *BlackMenu) getBalckMenuData() {
	fileName := this.blackFilePath + time.Now().Format("2006010215") + ".txt"
	f, err := os.Create(fileName)
	if err != nil {
		log.Error(err)
		return
	}
	defer f.Close()

	fn, err := os.Create(this.blackFileName)
	if err != nil {
		log.Error(err)
		return
	}
	defer fn.Close()
	for _, key := range this.ldb.Keys("*") {
		value := this.ldb.Get(key)
		f.WriteString(value + "\n")
		fn.WriteString(strings.TrimPrefix(key, this.keyprefix) + "\n")
	}
}

func (this *BlackMenu) getProAdverts() map[string]int {
	var advertMaps = make(map[string]int)
	this.rc_put.Send("SELECT", "0")
	prefixs := strings.Split(this.proprefix, ",")
	for _, pre := range prefixs {

		if infos, err := redis.Strings(this.rc_put.Do("SMEMBERS", pre)); err == nil {
			for _, v := range infos {
				advertMaps[v] = 1
			}
		} else {
			log.Fatal(err)
		}
	}
	return advertMaps
}

func (this *BlackMenu) setFilterAdvert() {
	var (
		db    = this.iniFile.Section("mongo").Key("db").String()
		table = "adreport"
		param = mongodb.MulQueryParam{}
		num   = 5
	)

	this.mp.Debug()
	param.DbName = db
	param.ColName = table
	param.Query = bson.M{}
	param.ChanSize = 1
	param.Fun = func(info map[string]interface{}) {
		if convert.ToInt(info["pv"]) >= num {
			if _, ok := this.provinceAdverts[info["aid"].(string)]; ok {
				if _, ok := info["click"]; !ok {
					ua := encrypt.DefaultBase64.Encode(info["tua"].(string))
					ad := info["ad"].(string)
					aid := info["aid"].(string)
					key := encrypt.DefaultMd5.Encode(ad + ua + aid)
					this.setDataToDb(key, aid+"_"+ad+"_"+ua)
				}
			}
		}
	}
	this.mp.Query(param)
	this.ldb.Flush()
}

func (this *BlackMenu) Do(c *cli.Context) {
	this.provinceAdverts = this.getProAdverts()
	this.setFilterAdvert()
	this.getBalckMenuData()
}

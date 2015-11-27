package blackad

import (
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/garyburd/redigo/redis"
	"github.com/ngaut/log"
	"github.com/qgweb/gopro/9xutool/common"
	"github.com/qgweb/gopro/lib/encrypt"
	"github.com/syndtr/goleveldb/leveldb"
	"gopkg.in/ini.v1"
	"gopkg.in/mgo.v2/bson"
	"io/ioutil"
	"runtime/debug"
	"strings"
)

// 广告黑名单生成
// 找出点击比较低的广告
type BlackMenu struct {
	iniFile         *ini.File
	mp              *common.MgoPool
	rc_put          redis.Conn
	blackprefix     string         // 黑名单
	provinceAdverts map[string]int // 浙江广告集合
	proprefix       string         // 浙江省对应的广告前缀
	data_path       string         // leveldb目录
	ldb             *leveldb.DB
}

// 获取monggo对象
func getMongoObj(inifile *ini.File) *common.MgoPool {
	mconf := &common.MgoConfig{}
	mconf.DBName = inifile.Section("mongo").Key("db").String()
	mconf.Host = inifile.Section("mongo").Key("host").String()
	mconf.Port = inifile.Section("mongo").Key("port").String()
	mconf.UserName = inifile.Section("mongo").Key("user").String()
	mconf.UserPwd = inifile.Section("mongo").Key("pwd").String()
	return common.NewMgoPool(mconf)
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
	rc.Send("SELECT", rdb)
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
			ur := &BlackMenu{}
			ur.iniFile = getConfig("black.conf")
			ur.mp = getMongoObj(ur.iniFile)
			ur.rc_put = getRedisObj("redis_put", ur.iniFile)
			ur.blackprefix = ur.iniFile.Section("default").Key("black_prefix").String()
			ur.proprefix = ur.iniFile.Section("default").Key("province_prefix").String()
			ur.data_path = ur.iniFile.Section("default").Key("data_path").String()
			ur.Do(c)
			ur.rc_put.Close()
			ur.ldb.Close()
		},
	}
}

func (this *BlackMenu) initLevelDb() {
	var err error
	this.ldb, err = leveldb.OpenFile(this.data_path, nil)
	if err != nil {
		log.Fatal(err)
		return
	}

	//清理原先的key
	iter := this.ldb.NewIterator(nil, nil)
	for iter.Next() {
		this.ldb.Delete(iter.Key(), nil)
	}
	iter.Release()
}

func (this *BlackMenu) setDataToDb(key string) {
	this.ldb.Put([]byte(key), []byte("1"), nil)
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
		db       = this.iniFile.Section("mongo").Key("db").String()
		table    = "adreport"
		num      = 5
		sess     = this.mp.Get()
		fadverts = make([]string, 0, len(this.provinceAdverts))
	)

	defer sess.Close()
	for k, _ := range this.provinceAdverts {
		fadverts = append(fadverts, k)
	}

	iter := sess.DB(db).C(table).Find(bson.M{"pv": bson.M{"$gte": num},
		"click": nil,
		"aid":   bson.M{"$in": fadverts}}).
		Select(bson.M{"ad": 1, "tua": 1, "aid": 1}).Iter()

	for {
		var info map[string]interface{}
		if !iter.Next(&info) {
			break
		}

		ua := encrypt.DefaultBase64.Encode(info["tua"].(string))
		ad := info["ad"].(string)
		aid := info["aid"].(string)
		key := encrypt.DefaultMd5.Encode(ad + ua + aid)
		this.setDataToDb(key)
	}

	iter.Close()
}

func (this *BlackMenu) Do(c *cli.Context) {
	this.initLevelDb()
	this.provinceAdverts = this.getProAdverts()
	this.setFilterAdvert()
}

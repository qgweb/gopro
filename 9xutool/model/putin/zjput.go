package putin

import (
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/garyburd/redigo/redis"
	"github.com/ngaut/log"
	"github.com/qgweb/gopro/9xutool/common"
	"gopkg.in/ini.v1"
	"gopkg.in/mgo.v2/bson"
	"io/ioutil"
	"os"
	"runtime/debug"
	"strings"
	"time"
)

// 浙江投放数据生成
type ZJPut struct {
	iniFile         *ini.File
	mp              *common.MgoPool
	rc_cache        redis.Conn
	rc_put          redis.Conn
	rc_dx_put       redis.Conn
	prefix          string
	proprefix       string                    // 浙江省对应的广告前缀
	tagMap0         map[string]map[string]int // cpc
	tagMap3         map[string]map[string]int // 横幅
	tagMap5         map[string]map[string]int // 医疗
	provinceAdverts map[string]int            // 浙江广告集合
	advertADS       map[string]map[string]int //广告对应ad集合
	tjprefix        string                    //统计prefix
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
		rc.Send("AUTH", rauth)
	}
	rc.Send("SELECT", rdb)
	return rc
}

func getConfig() *ini.File {
	// 获取配置文件
	filePath := common.GetBasePath() + "/conf/zp.conf"
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

func NewZheJiangPutCli() cli.Command {
	return cli.Command{
		Name:  "zhejiang_put",
		Usage: "生成浙江投放的数据",
		Action: func(c *cli.Context) {
			defer func() {
				if msg := recover(); msg != nil {
					log.Error(msg)
					debug.PrintStack()
				}
			}()
			ur := &ZJPut{}
			ur.iniFile = getConfig()
			ur.mp = getMongoObj(ur.iniFile)
			ur.rc_cache = getRedisObj("redis_cache", ur.iniFile)
			ur.rc_put = getRedisObj("redis_put", ur.iniFile)
			ur.rc_dx_put = getRedisObj("redis_dx_put", ur.iniFile)
			ur.prefix = bson.NewObjectId().Hex() + "_"
			ur.proprefix = ur.iniFile.Section("default").Key("province_prefix").String()
			ur.advertADS = make(map[string]map[string]int)
			ur.tjprefix = "advert_tj_zj_" + time.Now().Format("2006010215") + "_"
			ur.Do(c)
			ur.rc_cache.Close()
			ur.rc_put.Close()
			ur.rc_dx_put.Close()
		},
	}
}

// 获取省份的广告集合
func (this *ZJPut) getProAdverts() map[string]int {
	var advertMaps = make(map[string]int)
	this.rc_put.Send("SELECT", "0")
	if infos, err := redis.Strings(this.rc_put.Do("SMEMBERS", this.proprefix)); err == nil {
		for _, v := range infos {
			advertMaps[v] = 1
		}
		return advertMaps
	} else {
		log.Fatal(err)
		return nil
	}
}

// 获取投放中的标签广告
func (this *ZJPut) getTagsAdverts(key string) map[string]map[string]int {
	var adverMaps = make(map[string]map[string]int)
	this.rc_put.Send("SELECT", "0")
	if keys, err := redis.Strings(this.rc_put.Do("KEYS", key)); err == nil {
		for _, v := range keys {
			if _, ok := adverMaps[v]; !ok {
				adverMaps[v] = make(map[string]int)
			}
			if adlist, err := redis.Strings(this.rc_put.Do("SMEMBERS", v)); err == nil {
				for _, vv := range adlist {
					adverMaps[v][vv] = 1
				}
			}
		}
		return adverMaps
	} else {
		log.Fatal(err)
		return nil
	}
}

// 合并广告
func (this *ZJPut) merageAdverts(tag string) map[string]int {
	var keyPre0 = "TAGS_0_" + tag
	var keyPre3 = "TAGS_3_" + tag
	var info = make(map[string]int)
	if adverts, ok := this.tagMap0[keyPre0]; ok {
		for k, v := range adverts {
			if _, ok := this.provinceAdverts[k]; ok {
				info[k] = v
			}
		}
	}
	if adverts, ok := this.tagMap3[keyPre3]; ok {
		for k, v := range adverts {
			if _, ok := this.provinceAdverts[k]; ok {
				info[k] = v
			}
		}
	}
	return info
}

// 把广告信息写入投放系统
func (this *ZJPut) PutAdvertToRedis(ad string, advert string) {
	hashkey := "advert:" + advert
	this.rc_put.Send("HSET", ad, hashkey, advert)
	this.rc_put.Send("EXPIRE", ad, 3600)
}

// 把AD放入电信redis系统
func (this *ZJPut) PutDxSystem(ad string) {
	this.rc_dx_put.Send("SET", ad, "34")
}

// 把ad放入对应的广告集合里去
func (this *ZJPut) pushAdToAdvert(ad string, advertId string) {
	if _, ok := this.advertADS[advertId]; !ok {
		this.advertADS[advertId] = make(map[string]int)
	}
	this.advertADS[advertId][ad] = 1
}

func (this *ZJPut) flushDb() {
	this.rc_dx_put.Send("FLUSHDB")
}

// 医疗金融电商数据处理
func (this *ZJPut) Other() {
	var db = this.iniFile.Section("mongo").Key("db").String()
	var table = "useraction_put_big"
	var sess = this.mp.Get()
	var num = 0
	defer sess.Close()

	this.rc_put.Send("SELECT", "1")
	iter := sess.DB(db).C(table).Find(bson.M{}).
		Select(bson.M{"_id": 0, "AD": 1, "tag": 1}).Iter()
	for {
		var data map[string]interface{}
		if !iter.Next(&data) {
			break
		}
		num = num + 1
		log.Info(num)
		if tags, ok := data["tag"].([]interface{}); ok {
			ad := data["AD"].(string)
			for _, v := range tags {
				vm := v.(map[string]interface{})
				tagId := vm["tagId"].(string)
				for aid, _ := range this.merageAdverts(tagId) {
					//log.Warn(ad, aid)
					this.PutAdvertToRedis(ad, aid)
					this.pushAdToAdvert(ad, aid)
				}
			}
			this.PutDxSystem(ad)
		}
	}
	this.rc_put.Send("SELECT", "0")
}

// 域名
func (this *ZJPut) Domain() {
	var db = this.iniFile.Section("mongo").Key("db").String()
	var table = "urltrack_put"
	var sess = this.mp.Get()
	var num = 0
	defer sess.Close()

	this.rc_put.Send("SELECT", "1")
	iter := sess.DB(db).C(table).Find(bson.M{}).
		Select(bson.M{"_id": 0, "ad": 1, "cids": 1}).Iter()
	for {
		var data map[string]interface{}
		if !iter.Next(&data) {
			break
		}
		num = num + 1
		log.Info(num)
		if tags, ok := data["cids"].([]interface{}); ok {
			ad := data["ad"].(string)
			for _, v := range tags {
				vm := v.(map[string]interface{})
				tagId := vm["id"].(string)
				if ads, ok := this.tagMap5["TAGS_5_"+tagId]; ok {
					for aid, _ := range ads {
						//log.Warn(ad, aid)
						this.PutAdvertToRedis(ad, aid)
						this.pushAdToAdvert(ad, aid)
					}
				}
			}
			this.PutDxSystem(ad)
		}
	}
	this.rc_put.Send("SELECT", "0")
}

// 报错统计的数据
func (this *ZJPut) saveTjData() {
	var path = this.iniFile.Section("default").Key("data_path").String()
	this.rc_put.Send("SELECT", "3")
	for k, v := range this.advertADS {
		rk := this.tjprefix + k
		fname := path + "/" + rk + ".txt"
		if f, err := os.Create(fname); err == nil {
			for kk, _ := range v {
				f.WriteString(kk + "\n")
			}
			f.Close()
		}
		log.Debug(k, len(v))
		this.rc_put.Do("SET", rk, len(v))
	}
}

func (this *ZJPut) Do(c *cli.Context) {
	this.tagMap0 = this.getTagsAdverts("TAGS_0_*")
	this.tagMap3 = this.getTagsAdverts("TAGS_3_*")
	this.tagMap5 = this.getTagsAdverts("TAGS_5_*")
	this.provinceAdverts = this.getProAdverts()

	this.flushDb()
	this.Other()
	this.Domain()
	this.saveTjData()
}

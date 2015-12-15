package putin

import (
	"bufio"
	"github.com/codegangsta/cli"
	"github.com/garyburd/redigo/redis"
	"github.com/ngaut/log"
	"github.com/qgweb/gopro/9xutool/common"
	"github.com/qgweb/gopro/lib/convert"
	"github.com/qgweb/gopro/lib/encrypt"
	"gopkg.in/ini.v1"
	"gopkg.in/mgo.v2/bson"
	"io"
	"os"
	"runtime/debug"
	"strings"
	"time"
)

// 浙江投放数据生成
type JSPut struct {
	iniFile         *ini.File
	mp              *common.MgoPool
	rc_cache        redis.Conn
	rc_put          redis.Conn
	prefix          string
	proprefix       string                    // 浙江省对应的广告前缀
	blackprefix     string                    //黑名单
	tagMap0         map[string]map[string]int // cpc
	tagMap3         map[string]map[string]int // 横幅
	tagMap5         map[string]map[string]int // 医疗
	provinceAdverts map[string]int            // 浙江广告集合
	advertADS       map[string]map[string]int //广告对应ad集合
	tjprefix        string                    //统计prefix
	levelDataPath   string                    //level数据库目录
	blackMenus      map[string]int            // 黑名单
}

func NewJiangSuPutCli() cli.Command {
	return cli.Command{
		Name:  "jiangsu_put",
		Usage: "生成江苏投放的数据",
		Action: func(c *cli.Context) {
			defer func() {
				if msg := recover(); msg != nil {
					log.Error(msg)
					debug.PrintStack()
				}
			}()
			ur := &JSPut{}
			ur.iniFile = getConfig("js.conf")
			ur.mp = getMongoObj(ur.iniFile)
			ur.rc_cache = getRedisObj("redis_cache", ur.iniFile)
			ur.rc_put = getRedisObj("redis_put", ur.iniFile)
			ur.prefix = bson.NewObjectId().Hex() + "_"
			ur.proprefix = ur.iniFile.Section("default").Key("province_prefix").String()
			ur.blackprefix = ur.iniFile.Section("default").Key("black_prefix").String()
			ur.levelDataPath = ur.iniFile.Section("default").Key("level_data_path").String()
			ur.advertADS = make(map[string]map[string]int)
			ur.tjprefix = "advert_tj_js_" + time.Now().Format("2006010215") + "_"
			ur.Do(c)
			ur.rc_cache.Close()
			ur.rc_put.Close()
		},
	}
}

// 初始化leveldb
func (this *JSPut) initLevelDb() {
	this.blackMenus = make(map[string]int)
	f, err := os.Open(this.levelDataPath)
	if err != nil {
		log.Error(err)
		return
	}
	defer f.Close()

	br := bufio.NewReader(f)
	for {
		l, err := br.ReadString('\n')
		if err == io.EOF {
			break
		}

		this.blackMenus[l] = 1
	}
}

// 判断key是否存在
func (this *JSPut) existKey(key string) bool {
	if _, ok := this.blackMenus[key]; ok {
		return true
	}
	return false
}

// 过滤黑名单中的广告
func (this *JSPut) filterAdvert(key string, list map[string]int) {
	for aid, _ := range list {
		key := encrypt.DefaultMd5.Encode(key + aid)
		if this.existKey(key) {
			delete(list, aid)
		}
	}
}

// 获取省份的广告集合
func (this *JSPut) getProAdverts() map[string]int {
	var advertMaps = make(map[string]int)
	this.rc_put.Do("SELECT", "0")
	if infos, err := redis.Strings(this.rc_put.Do("SMEMBERS", this.proprefix)); err == nil {
		for _, v := range infos {
			advertMaps[v] = 1
		}
	} else {
		log.Fatal(err)
	}
	return advertMaps
}

// 获取投放中的标签广告
func (this *JSPut) getTagsAdverts(key string) map[string]map[string]int {
	var adverMaps = make(map[string]map[string]int)
	this.rc_put.Do("SELECT", "0")
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
func (this *JSPut) merageAdverts(tag string) map[string]int {
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

func (this *JSPut) merageAdverts2(tag string) map[string]int {
	var keyPre5 = "TAGS_5_" + tag
	var info = make(map[string]int)
	if adverts, ok := this.tagMap5[keyPre5]; ok {
		for k, v := range adverts {
			if _, ok := this.provinceAdverts[k]; ok {
				info[k] = v
			}
		}
	}
	return info
}

// 把广告信息写入投放系统
func (this *JSPut) PutAdvertToRedis(ad string, ua string, advert string) {
	hashkey := "advert:" + advert
	key := ad
	if strings.ToLower(ua) != "ua" {
		key = encrypt.DefaultMd5.Encode(ad + "_" + ua)
	}
	this.rc_put.Do("HSET", key, hashkey, advert)
	this.rc_put.Do("EXPIRE", key, 3600)
}

// 把AD放入缓存redis系统
func (this *JSPut) PutDxSystem(ad string) {
	this.rc_cache.Do("SET", this.prefix+ad, "1")
}

// 把ad放入对应的广告集合里去
func (this *JSPut) pushAdToAdvert(ad string, ua string, advertId string) {
	if _, ok := this.advertADS[advertId]; !ok {
		this.advertADS[advertId] = make(map[string]int)
	}
	key := ad + "_" + ua
	this.advertADS[advertId][key] = 1
}

// 清空缓存key
func (this *JSPut) emptyDb() {
	if keys, err := redis.Strings(this.rc_cache.Do("KEYS", this.prefix+"*")); err == nil {
		for _, v := range keys {
			this.rc_cache.Do("DEL", v)
		}
	}
}

// 医疗金融电商数据处理
func (this *JSPut) Other(query bson.M) {
	var db = this.iniFile.Section("mongo").Key("db").String()
	var table = "useraction_jiangsu"
	var sess = this.mp.Get()
	var num = 0
	defer sess.Close()

	this.rc_put.Do("SELECT", "1")
	iter := sess.DB(db).C(table).Find(query).
		Select(bson.M{"_id": 0, "AD": 1, "UA": 1, "tag": 1}).Iter()
	for {
		var data map[string]interface{}
		if !iter.Next(&data) {
			break
		}
		num = num + 1
		if num%10000 == 0 {
			log.Debug(num)
		}

		if tags, ok := data["tag"].([]interface{}); ok {
			ad := data["AD"].(string)
			ua := data["UA"].(string)
			for _, v := range tags {
				vm := v.(map[string]interface{})
				tagId := vm["tagId"].(string)
				piadverts := this.merageAdverts(tagId)
				this.filterAdvert(ad+ua, piadverts)
				if len(piadverts) > 0 {
					this.PutDxSystem(ad)
				}
				for aid, _ := range piadverts {
					//log.Warn(ad, aid)
					this.PutAdvertToRedis(ad, ua, aid)
					this.pushAdToAdvert(ad, ua, aid)
				}
			}
		}
	}
	this.rc_put.Do("SELECT", "0")
	log.Info(num)
}

// 域名
func (this *JSPut) Domain(query bson.M) {
	var db = this.iniFile.Section("mongo").Key("db").String()
	var table = "urltrack_jiangsu_" + time.Now().Format("200601")
	var sess = this.mp.Get()
	var num = 0
	defer sess.Close()

	this.rc_put.Do("SELECT", "1")
	iter := sess.DB(db).C(table).Find(query).
		Select(bson.M{"_id": 0, "ad": 1, "ua": 1, "cids": 1}).Iter()
	log.Info(query, table)
	for {
		var data map[string]interface{}
		if !iter.Next(&data) {
			break
		}
		num = num + 1
		if num%10000 == 0 {
			log.Debug(num)
		}
		if tags, ok := data["cids"].([]interface{}); ok {
			ad := data["ad"].(string)
			ua := data["ua"].(string)
			for _, v := range tags {
				vm := v.(map[string]interface{})
				tagId := vm["id"].(string)
				piadverts := this.merageAdverts2(tagId)
				log.Info(piadverts)
				this.filterAdvert(ad+ua, piadverts)
				log.Info(piadverts)
				if len(piadverts) > 0 {
					this.PutDxSystem(ad)
				}
				for aid, _ := range piadverts {
					//log.Warn(ad, aid)
					this.PutAdvertToRedis(ad, ua, aid)
					this.pushAdToAdvert(ad, ua, aid)
				}
			}
		}
	}
	this.rc_put.Do("SELECT", "0")
	log.Info(num)
}

// 报错统计的数据
func (this *JSPut) saveTjData() {
	var path = this.iniFile.Section("default").Key("data_path").String()
	this.rc_put.Do("SELECT", "3")
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

// 保存投放ad到文件
func (this *JSPut) savePutData() {
	var path = this.iniFile.Section("default").Key("put_path").String()
	rk := time.Now().Add(-time.Hour).Format("2006010215")
	fname := path + "/" + rk + ".txt"
	f, err := os.Create(fname)
	if err != nil {
		log.Error(err)
		return
	}
	if list, err := redis.Strings(this.rc_cache.Do("KEYS", this.prefix+"*")); err == nil {
		for _, v := range list {
			f.WriteString(strings.Split(v, "_")[1] + "\n")
		}
		f.Close()
	}
}

func (this *JSPut) Test() {
	this.rc_put.Do("SELECT", "5")
	bt := time.Now()
	for i := 0; i <= 100000; i++ {
		this.rc_put.Do("SET", i, i)
	}
	log.Debug(time.Now().Sub(bt).Seconds())
}

func (this *JSPut) Do(c *cli.Context) {
	var (
		now    = time.Now()
		now1   = now.Add(-time.Second * time.Duration(now.Second())).Add(-time.Minute * time.Duration(now.Minute()))
		eghour = convert.ToString(now1.Add(-time.Hour).Unix())
		bghour = convert.ToString(now1.Add(-time.Duration(time.Hour * 2)).Unix())
	)
	//this.initLevelDb()
	this.tagMap0 = this.getTagsAdverts("TAGS_0_*")
	this.tagMap3 = this.getTagsAdverts("TAGS_3_*")
	this.tagMap5 = this.getTagsAdverts("TAGS_5_*")
	this.provinceAdverts = this.getProAdverts()

	this.Other(bson.M{"timestamp": bson.M{"$gte": bghour, "$lte": eghour}})
	this.Domain(bson.M{"timestamp": bson.M{"$gte": bghour, "$lte": eghour}})
	this.saveTjData()
	this.savePutData()
	this.emptyDb()
}

package putin

import (
	"bufio"
	"github.com/codegangsta/cli"
	"github.com/garyburd/redigo/redis"
	"github.com/ngaut/log"
	"github.com/qgweb/gopro/9xutool/common"
	"github.com/qgweb/gopro/lib/cache"
	"github.com/qgweb/gopro/lib/encrypt"
	"github.com/qgweb/gopro/lib/mongodb"
	"gopkg.in/ini.v1"
	"gopkg.in/mgo.v2/bson"
	"io"
	"os"
	"os/exec"
	"runtime/debug"
	"strings"
	"sync"
	"time"
)

// 浙江投放数据生成
type JSPut struct {
	iniFile         *ini.File
	mp              *mongodb.DialContext
	mp_tj           *mongodb.DialContext
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
	blackFileName   string                    //黑名单文件
	blackMenus      map[string]int            // 黑名单
	ldb             *cache.LevelDBCache       //ldb缓存类
	mux             sync.Mutex
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
			var err error
			ur := &JSPut{}
			ur.iniFile = getConfig("js.conf")
			if ur.mp, err = ur.getMongoObj(ur.iniFile, "mongo"); err != nil {
				log.Fatal(err)
				return
			}
			if ur.mp_tj, err = ur.getMongoObj(ur.iniFile, "mongo-tj"); err != nil {
				log.Fatal(err)
				return
			}
			ur.mp.Debug()
			ur.rc_put = getRedisObj("redis_put", ur.iniFile)
			ur.prefix = bson.NewObjectId().Hex() + "_"
			ur.proprefix = ur.iniFile.Section("default").Key("province_prefix").String()
			ur.blackprefix = ur.iniFile.Section("default").Key("black_prefix").String()
			ur.blackFileName = ur.iniFile.Section("default").Key("black_data_file").String()
			ur.advertADS = make(map[string]map[string]int)
			ur.tjprefix = "advert_tj_js_" + time.Now().Format("2006010215") + "_"
			ur.ldb = ur.initLevelDb()
			ur.Do(c)
			ur.mp.Close()
			ur.mp_tj.Close()
			ur.ldb.Close()
			ur.rc_put.Close()
		},
	}
}

// 获取monggo对象
func (this *JSPut) getMongoObj(inifile *ini.File, section string) (*mongodb.DialContext, error) {
	mconf := mongodb.MgoConfig{}
	mconf.DBName = inifile.Section(section).Key("db").String()
	mconf.Host = inifile.Section(section).Key("host").String()
	mconf.Port = inifile.Section(section).Key("port").String()
	mconf.UserName = inifile.Section(section).Key("user").String()
	mconf.UserPwd = inifile.Section(section).Key("pwd").String()
	return mongodb.Dial(mongodb.GetLinkUrl(mconf), mongodb.GetCpuSessionNum())
}

func (this *JSPut) initLevelDb() *cache.LevelDBCache {
	path := common.GetBasePath() + "/" + bson.NewObjectId().Hex()
	if ldb, err := cache.NewLevelDbCache(path); err == nil {
		return ldb
	}
	return nil
}

// 初始化黑名单
func (this *JSPut) initBlackMenu() {
	this.blackMenus = make(map[string]int)
	f, err := os.Open(this.blackFileName)
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

// 把广告信息写入缓存系统
func (this *JSPut) PutAdvertToCache(ad string, ua string, advert string) {
	hashkey := "advert:" + advert
	key := ad
	if strings.ToLower(ua) != "ua" {
		key = encrypt.DefaultMd5.Encode(ad + "_" + ua)
	}
	this.ldb.HSet("adua_"+key, hashkey, advert)
	this.ldb.HSet("adua_keys", "adua_"+key, "1")
}

// 把广告信息写入投放系统
func (this *JSPut) PutAdvertToRedis() {
	this.rc_put.Do("SELECT", "1")
	if keys, err := this.ldb.HGetAllKeys("adua_keys"); err == nil {
		bt := time.Now()
		var count = 0
		var hour float64
		for _, key := range keys {
			hkey := strings.TrimPrefix(key, "adua_")
			eflag := 0
			bbt := time.Now()
			if adverts, err := this.ldb.HGetAllValue(key); err == nil {
				hour += time.Now().Sub(bbt).Seconds()
				for _, advert := range adverts {
					count++
					this.rc_put.Send("HSET", hkey, "advert:"+advert, advert)
					if eflag = eflag + 1; eflag == 1 {
						this.rc_put.Send("EXPIRE", hkey, 3600)
					}
				}
			}
		}
		log.Info(count, time.Now().Sub(bt).Seconds())
		this.rc_put.Flush()
		this.rc_put.Receive()
		log.Info(count, hour)
	}
	this.rc_put.Do("SELECT", "0")
}

// 把AD放入缓存中
func (this *JSPut) PutAdToCache(ad string) {
	this.ldb.HSet("sad", ad, "1")
}

// 把ad放入对应的广告集合里去
func (this *JSPut) pushAdToAdvert(ad string, ua string, advertId string) {
	defer func() {
		if msg := recover(); msg != nil {
			log.Info(msg)
		}
	}()
	this.mux.Lock()
	defer this.mux.Unlock()
	if _, ok := this.advertADS[advertId]; !ok {
		this.advertADS[advertId] = make(map[string]int)
	}
	key := ad + "_" + ua
	this.advertADS[advertId][key] = 1

}

// 医疗金融电商数据处理
func (this *JSPut) Other(query bson.M) {
	var (
		table = "useraction_jiangsu"
		db    = this.iniFile.Section("mongo").Key("db").String()
		param = mongodb.MulQueryParam{}
	)
	log.Info(query)
	param.DbName = db
	param.ColName = table
	param.Query = query
	param.Fun = func(data map[string]interface{}) {
		if tags, ok := data["tag"].([]interface{}); ok {
			ad := data["AD"].(string)
			ua := data["UA"].(string)

			for _, v := range tags {
				vm := v.(map[string]interface{})
				tagId := vm["tagId"].(string)
				piadverts := this.merageAdverts(tagId)
				this.filterAdvert(ad+ua, piadverts)

				if len(piadverts) > 0 {
					this.PutAdToCache(ad)
				}

				for aid, _ := range piadverts {
					this.PutAdvertToCache(ad, ua, aid)
					this.pushAdToAdvert(ad, ua, aid)
				}
			}
		}
	}

	this.mp.Query(param)
}

// 域名
func (this *JSPut) Domain(query bson.M) {
	var (
		db    = this.iniFile.Section("mongo").Key("db").String()
		table = "urltrack_jiangsu_" + time.Now().Format("200601")
		param = mongodb.MulQueryParam{}
	)
	param.DbName = db
	param.ColName = table
	param.Query = query
	param.Fun = func(data map[string]interface{}) {
		if tags, ok := data["cids"].([]interface{}); ok {
			ad := data["ad"].(string)
			ua := data["ua"].(string)
			for _, v := range tags {
				vm := v.(map[string]interface{})
				tagId := vm["id"].(string)
				piadverts := this.merageAdverts2(tagId)
				this.filterAdvert(ad+ua, piadverts)
				if len(piadverts) > 0 {
					this.PutAdToCache(ad)
				}
				for aid, _ := range piadverts {
					this.PutAdvertToCache(ad, ua, aid)
					this.pushAdToAdvert(ad, ua, aid)
				}
			}
		}
	}
	this.mp.Query(param)
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
	var path2 = this.iniFile.Section("default").Key("put_path2").String()
	var ftp = this.iniFile.Section("default").Key("ftp_bash").String()

	rk := time.Now().Add(-time.Hour).Format("2006010215")
	fname := path + "/" + rk + ".txt"
	fname2 := path2 + "/tag_" + rk + ".txt"

	f, err := os.Create(fname)
	if err != nil {
		log.Error(err)
		return
	}

	f1, err := os.Create(fname2)
	if err != nil {
		log.Error(err)
		return
	}

	if ads, err := this.ldb.HGetAllKeys("sad"); err == nil {
		log.Info("总ad数", len(ads))
		for _, ad := range ads {
			//f.WriteString(ad + "\n")
			f1.WriteString(ad + "\n")
		}
		f.Close()
		f1.Close()
	}

	//提交ftp
	cmd := exec.Command(ftp, "tag_"+rk+".txt")
	str, err := cmd.Output()
	log.Info(string(str), err)
}

// 点击的白名单
func (this *JSPut) GetClickWhiteMenu() {
	var (
		sess    = this.mp_tj.Ref()
		db      = this.iniFile.Section("mongo-tj").Key("db").String()
		table   = "adreport_click"
		adverts = make([]string, 0, len(this.provinceAdverts))
	)
	defer this.mp_tj.UnRef(sess)

	for k, _ := range this.provinceAdverts {
		adverts = append(adverts, k)
	}

	iter := sess.DB(db).C(table).Find(bson.M{"aid": bson.M{"$in": adverts}}).Iter()
	for {
		var info map[string]interface{}
		if !iter.Next(&info) {
			break
		}
		this.PutAdvertToCache(info["ad"].(string), info["ua"].(string), info["aid"].(string))
		this.pushAdToAdvert(info["ad"].(string), info["ua"].(string), info["aid"].(string))
		this.PutAdToCache(info["ad"].(string))
	}
	iter.Close()
}

func (this *JSPut) Do(c *cli.Context) {
	var (
		eghour = common.GetHourTimestamp(-1)
		bghour = common.GetHourTimestamp(-2)
	)

	this.initBlackMenu()
	this.tagMap0 = this.getTagsAdverts("TAGS_0_*")
	this.tagMap3 = this.getTagsAdverts("TAGS_3_*")
	this.tagMap5 = this.getTagsAdverts("TAGS_5_*")
	this.provinceAdverts = this.getProAdverts()

	this.Other(bson.M{"timestamp": bson.M{"$gte": bghour, "$lte": eghour}})
	this.Domain(bson.M{"timestamp": bson.M{"$gte": bghour, "$lte": eghour}})
	this.GetClickWhiteMenu()
	this.PutAdvertToRedis()
	this.saveTjData()
	this.savePutData()
}

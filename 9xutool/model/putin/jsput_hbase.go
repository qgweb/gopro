package putin

import (
	"bufio"
	"github.com/codegangsta/cli"
	"github.com/garyburd/redigo/redis"
	"github.com/ngaut/log"
	"github.com/qgweb/go-hbase"
	"github.com/qgweb/gopro/9xutool/common"
	"github.com/qgweb/gopro/lib/encrypt"
	"github.com/qgweb/gopro/lib/mongodb"
	"github.com/qgweb/gopro/lib/rediscache"
	"github.com/qgweb/new/lib/timestamp"
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
type JSPutHbase struct {
	iniFile         *ini.File
	hc              hbase.HBaseClient
	mp_tj           *mongodb.DialContext
	rc_put          redis.Conn
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
	ldb             *rediscache.MemCache      //ldb缓存类
	mux             sync.Mutex
	keyprefix       string
	coxarea         string // cox对应区域
	areaMap         map[string]string
}

func NewJiangSuPutHbaseCli() cli.Command {
	return cli.Command{
		Name:  "jiangsu_put_hbase",
		Usage: "生成江苏投放的数据",
		Action: func(c *cli.Context) {
			defer func() {
				if msg := recover(); msg != nil {
					log.Error(msg)
					debug.PrintStack()
				}
			}()
			var err error
			ur := &JSPutHbase{}
			ur.iniFile = getConfig("js.conf")
			if ur.hc, err = common.GetHbaseObj("js.conf", "hbase"); err != nil {
				log.Fatal(err)
				return
			}
			if ur.mp_tj, err = ur.getMongoObj(ur.iniFile, "mongo-tj"); err != nil {
				log.Fatal(err)
				return
			}

			ur.rc_put = getRedisObj("redis_put", ur.iniFile)
			ur.proprefix = ur.iniFile.Section("default").Key("province_prefix").String()
			ur.blackprefix = ur.iniFile.Section("default").Key("black_prefix").String()
			ur.blackFileName = ur.iniFile.Section("default").Key("black_data_file").String()
			ur.advertADS = make(map[string]map[string]int)
			ur.tjprefix = "advert_tj_js_" + time.Now().Format("2006010215") + "_"
			ur.keyprefix = mongodb.GetObjectId() + "_"
			ur.coxarea = ur.iniFile.Section("default").Key("cox_area").String()
			ur.initLevelDb()
			ur.Do(c)
			ur.mp_tj.Close()
			ur.ldb.Clean(ur.keyprefix)
			ur.ldb.Clean(ur.keyprefix)
			ur.ldb.Close()
			ur.rc_put.Close()
		},
	}
}

// 获取monggo对象
func (this *JSPutHbase) getMongoObj(inifile *ini.File, section string) (*mongodb.DialContext, error) {
	mconf := mongodb.MgoConfig{}
	mconf.DBName = inifile.Section(section).Key("db").String()
	mconf.Host = inifile.Section(section).Key("host").String()
	mconf.Port = inifile.Section(section).Key("port").String()
	mconf.UserName = inifile.Section(section).Key("user").String()
	mconf.UserPwd = inifile.Section(section).Key("pwd").String()
	return mongodb.Dial(mongodb.GetLinkUrl(mconf), mongodb.GetCpuSessionNum())
}

func (this *JSPutHbase) initLevelDb() {
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

// 初始化黑名单
func (this *JSPutHbase) initBlackMenu() {
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

func (this *JSPutHbase) initAreaMap() {
	this.areaMap = make(map[string]string)
	f, err := os.Open(this.coxarea)
	if err != nil {
		log.Error(err)
		return
	}
	defer f.Close()

	bi := bufio.NewReader(f)
	for {
		line, err := bi.ReadString('\n')
		if err == io.EOF {
			break
		}

		//0006f21d119b032d59acc3c2b90f10624eeaebe8,511
		info := strings.Split(line, ",")
		if len(info) != 2 {
			continue
		}

		this.areaMap[info[0]] = strings.TrimSpace(info[1])
	}
	log.Info(len(this.areaMap))
}

// 判断key是否存在
func (this *JSPutHbase) existKey(key string) bool {
	if _, ok := this.blackMenus[key]; ok {
		return true
	}
	return false
}

// 过滤黑名单中的广告
func (this *JSPutHbase) filterAdvert(key string, list map[string]int) {
	for aid, _ := range list {
		key := encrypt.DefaultMd5.Encode(key + aid)
		if this.existKey(key) {
			delete(list, aid)
		}
	}
}

// 获取省份的广告集合
func (this *JSPutHbase) getProAdverts() map[string]int {
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
func (this *JSPutHbase) getTagsAdverts(key string) map[string]map[string]int {
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
func (this *JSPutHbase) merageAdverts(tag string) map[string]int {
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

func (this *JSPutHbase) merageAdverts2(tag string) map[string]int {
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
func (this *JSPutHbase) PutAdvertToCache(ad string, ua string, advert string) {
	hashkey := "advert:" + advert
	key := ad
	if strings.ToLower(ua) != "ua" {
		key = encrypt.DefaultMd5.Encode(ad + "_" + ua)
	}

	this.mux.Lock()
	this.ldb.HSet(this.keyprefix+key, hashkey, advert)
	this.mux.Unlock()
}

// 把广告信息写入投放系统
func (this *JSPutHbase) PutAdvertToRedis() {
	this.rc_put.Do("SELECT", "1")
	keys := this.ldb.Keys(this.keyprefix + "*")
	bt := time.Now()
	var count = 0
	var hour float64
	for _, key := range keys {
		hkey := strings.TrimPrefix(key, this.keyprefix)
		eflag := 0
		bbt := time.Now()
		adverts := this.ldb.HGetAllValue(key)
		hour += time.Now().Sub(bbt).Seconds()
		for _, advert := range adverts {
			count++
			this.rc_put.Send("HSET", hkey, "advert:"+advert, advert)
			if eflag = eflag + 1; eflag == 1 {
				this.rc_put.Send("EXPIRE", hkey, 5400)
			}
		}

	}
	log.Info(count, time.Now().Sub(bt).Seconds())
	this.rc_put.Flush()
	log.Info(count, hour)
	this.rc_put.Do("SELECT", "0")
}

// 把AD放入缓存中
func (this *JSPutHbase) PutAdToCache(ad string) {
	this.mux.Lock()
	this.ldb.HSet(this.keyprefix+"sad", ad, "1")
	this.mux.Unlock()
}

// 把ad放入对应的广告集合里去
func (this *JSPutHbase) pushAdToAdvert(ad string, ua string, advertId string) {
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
func (this *JSPutHbase) Other() {
	var (
		table = []byte("jiangsu_useraction_" + timestamp.GetMonthTimestamp(0))
		btime = []byte(timestamp.GetHourTimestamp(-1))
		etime = []byte(timestamp.GetHourTimestamp(0))
	)

	sc := hbase.NewScan(table, 10000, this.hc)
	sc.StartRow = btime
	sc.StopRow = etime

	for {
		info := sc.Next()
		if info == nil {
			break
		}

		ad := string(info.Columns["base:ad"].Value)
		ua := string(info.Columns["base:ua"].Value)

		for _, v := range info.Columns {
			if string(v.Family) == "cids" {
				tagId := string(v.Qual)
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
	this.ldb.Flush()
	sc.Close()
}

// 域名
func (this *JSPutHbase) Domain() {
	var (
		table = []byte("jiangsu_urltrack_" + timestamp.GetMonthTimestamp(0))
		btime = []byte(timestamp.GetHourTimestamp(-1))
		etime = []byte(timestamp.GetHourTimestamp(0))
	)

	sc := hbase.NewScan(table, 10000, this.hc)
	sc.StartRow = btime
	sc.StopRow = etime

	for {
		info := sc.Next()
		if info == nil {
			break
		}

		ad := string(info.Columns["base:ad"].Value)
		ua := string(info.Columns["base:ua"].Value)

		for _, v := range info.Columns {
			if string(v.Family) == "cids" {
				tagId := string(v.Qual)
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
	this.ldb.Flush()
	sc.Close()
}

// 报错统计的数据
func (this *JSPutHbase) saveTjData() {
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
func (this *JSPutHbase) savePutData() {
	var path1 = this.iniFile.Section("default").Key("put_path2").String()
	var path2 = this.iniFile.Section("default").Key("put_path3").String()
	var ftp1 = this.iniFile.Section("default").Key("ftp_bash1").String()
	var ftp2 = this.iniFile.Section("default").Key("ftp_bash2").String()

	rk1 := time.Now().Add(-time.Hour).Format("2006010215")
	rk2 := "account.10046.sha1." + time.Now().Add(-time.Hour).Format("200601021504")

	fname1 := path1 + "/tag_" + rk1 + ".txt"
	fname2 := path2 + "/" + rk2

	f1, err := os.Create(fname1)
	if err != nil {
		log.Error(err)
		return
	}
	f2, err := os.Create(fname2)
	if err != nil {
		log.Error(err)
		return
	}

	ads := this.ldb.HGetAllKeys(this.keyprefix + "sad")
	log.Info(len(ads))
	log.Info("总ad数", len(ads))
	for _, ad := range ads {
		f1.WriteString(ad + "\n")
		if v, ok := this.areaMap[ad]; ok {
			f2.WriteString(ad + "," + v + "\n")
		}
	}
	f1.Close()
	f2.Close()

	//提交ftp
	cmd := exec.Command(ftp1, "tag_"+rk1+".txt")
	str, err := cmd.Output()
	log.Info(string(str), err)
	log.Info(ftp2, rk2)
	cmd = exec.Command(ftp2, rk2)
	str, err = cmd.Output()
	log.Info(string(str), err)
}

// 点击的白名单
func (this *JSPutHbase) GetClickWhiteMenu() {
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

func (this *JSPutHbase) Do(c *cli.Context) {
	log.Info(this.keyprefix)
	this.initBlackMenu()
	this.initAreaMap()
	this.tagMap0 = this.getTagsAdverts("TAGS_0_*")
	this.tagMap3 = this.getTagsAdverts("TAGS_3_*")
	this.tagMap5 = this.getTagsAdverts("TAGS_5_*")
	this.provinceAdverts = this.getProAdverts()

	this.Other()
	this.Domain()
	//this.GetClickWhiteMenu()
	this.PutAdvertToRedis()
	this.saveTjData()
	this.savePutData()
}
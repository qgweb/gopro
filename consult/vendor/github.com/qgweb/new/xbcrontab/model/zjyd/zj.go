package zjyd

import (
	"fmt"
	"strings"
	"time"

	"github.com/codegangsta/cli"
	"github.com/ngaut/log"
	"github.com/qgweb/new/lib/convert"
	"github.com/qgweb/new/lib/dbfactory"
	"github.com/qgweb/new/lib/encrypt"
	"github.com/qgweb/new/lib/timestamp"
	"github.com/qgweb/new/xbcrontab/lib"
)

func CliPutData() cli.Command {
	return cli.Command{
		Name:   "zhejiang_yd_put",
		Usage:  "浙江移动投放数据",
		Action: putRun,
	}
}

func putRun(ctx *cli.Context) {
	jp := NewZjPut()
	jp.Run()
	jp.Clean()
}

func NewZjPut() *ZjPut {
	var zj = &ZjPut{}
	zj.kf = dbfactory.NewKVFile(fmt.Sprintf("./%s.txt", convert.ToString(time.Now().Unix())))
	zj.putTags = make(map[string]map[string]int)
	zj.Timestamp = timestamp.GetHourTimestamp(-1)
	zj.initPutAdverts()
	zj.initPutTags("TAGS_3*", "tb_phone_", "mg_phone_")
	zj.initPutTags("TAGS_5*", "url_phone_", "")
	return zj
}

// 浙江投放
type ZjPut struct {
	kf         *dbfactory.KVFile
	putAdverts map[string]int
	putTags    map[string]map[string]int
	Timestamp  string
}

// 初始化需要投放的广告
func (this *ZjPut) initPutAdverts() {
	rdb, err := lib.GetRedisObj()
	if err != nil {
		log.Fatal(err)
	}
	rdb.SelectDb("0")
	this.putAdverts = make(map[string]int)
	alist := rdb.SMembers(lib.GetConfVal("zhejiang::province_web_prefix"))
	for _, v := range alist {
		this.putAdverts[v] = 1
	}
	rdb.Close()
}

// 初始化投放标签
func (this *ZjPut) initPutTags(tagkey string, prefix1 string, prefix2 string) {
	rdb, err := lib.GetRedisObj()
	if err != nil {
		log.Fatal(err)
	}
	rdb.SelectDb("0")
	for _, key := range rdb.Keys(tagkey) {
		rkey := strings.TrimPrefix(key, strings.TrimSuffix(tagkey, "*")+"_")
		if lib.IsMongo(rkey) {
			rkey = prefix2 + rkey
		} else {
			rkey = prefix1 + rkey
		}
		if _, ok := this.putTags[rkey]; !ok {
			this.putTags[rkey] = make(map[string]int)
		}
		for _, aid := range rdb.SMembers(key) {
			if _, ok := this.putAdverts[aid]; ok {
				this.putTags[rkey][aid] = 1
			}
		}
	}
}

// 无限数据获取
func (this *ZjPut) wapData(out chan interface{}, in chan int8) {
	var datacount = 0
	defer func() {
		// 统计数据 jiangsu_put , url_1461016800, 11111
		lib.StatisticsData("dsource_stats", "zj_"+this.Timestamp+"_urlphone",
			convert.ToString(datacount), "")
	}()

	fname := "zhejiang_url_phone_" + this.Timestamp
	if err := lib.GetFdbData(fname, func(val string) {
		if v := lib.AddPrefix(val, "url_phone_"); v != "" {
			datacount++
			out <- v
		}
	}); err != nil {
		in <- 1
		return
	}
	log.Info("无线ok")
	in <- 1
}

// 标签数据统计
func (this *ZjPut) tagDataStats() {
	fname := convert.ToString(time.Now().UnixNano()) + "_"
	this.kf.IDAdUaSet(fname, func(info map[string]int) {
		for k, v := range info {
			tagid := strings.TrimPrefix(k, fname)
			tagids := strings.Split(tagid, "_")
			// 标签统计数据 tags_stats , url_1461016800, 11111
			lib.StatisticsData("tags_stats", fmt.Sprintf("zj_%s_%s_%s", this.Timestamp, tagids[0], tagids[1]),
				convert.ToString(v), "incr")
		}
	}, true)
}

// 过滤数据
func (this *ZjPut) filterData() {
	this.kf.Filter(func(info dbfactory.AdUaAdverts) (string, bool) {
		var advertIds = make(map[string]int)
		for tagid := range info.AId {
			// 标签
			if v, ok := this.putTags[tagid]; ok {
				for aid := range v {
					advertIds[aid] = 1
				}
			}
		}
		var aids = make([]string, 0, len(advertIds))
		for k := range advertIds {
			aids = append(aids, k)
		}
		if len(aids) != 0 {
			return fmt.Sprintf("%s\t%s\t%s", info.Ad, info.UA, strings.Join(aids, ",")), true
		}
		return "", false
	})
}

// 保存广告对应的ad，ua
func (this *ZjPut) saveAdvertSet() {
	tname := "advert_tj_zj_" + this.Timestamp + "_"
	fname := lib.GetConfVal("zhejiang::data_path") + tname
	this.kf.IDAdUaSet(fname, func(info map[string]int) {
		tm := this.Timestamp
		for k, v := range info {
			aid := strings.TrimPrefix(k, tname)
			// 广告数量统计数据 advert_stats , zj_1461016800_1111, 11111
			lib.StatisticsData("advert_stats", fmt.Sprintf("zj_%s_%s", tm, aid),
				convert.ToString(v), "incr")
		}
	}, false)
}

// 保存投放轨迹到投放系统
func (this *ZjPut) saveTraceToPutSys() {
	rdb, err := lib.GetRedisObj()
	if err != nil {
		log.Error("redis连接失败", err)
		return
	}
	rdb.SelectDb("1")
	adcount := 0
	this.kf.AdUaIdsSet(func(ad string, ua string, aids map[string]int8) {
		key := ad
		if ua != "ua" {
			key = encrypt.DefaultMd5.Encode(ad + "_" + ua)
		}
		for aid, _ := range aids {
			rdb.HSet(key, "advert:"+aid, aid)
		}
		rdb.Expire(key, 5400)
		adcount++
	})
	rdb.Flush()
	rdb.Close()
	// 广告数量统计数据 put_stats , Zj_1461016800, 11111
	lib.StatisticsData("put_stats", fmt.Sprintf("zj_%s", this.Timestamp),
		convert.ToString(adcount), "incr")
}

// 保存投放轨迹到电信redis
func (this *ZjPut) saveTraceToDianxin() {
	var (
		db      = lib.GetConfVal("zhejiang::dx_redis_db")
		pwd     = lib.GetConfVal("zhejiang::dx_redis_pwd")
		adcount = 0
	)

	rdb, err := lib.GetZJDxRedisObj()
	if err != nil {
		log.Error("redis连接失败", err)
		return
	}
	rdb.Auth(pwd)
	rdb.SelectDb(db)

	// ua默认md5加密
	this.kf.AdUaIdsSet(func(ad string, ua string, ids map[string]int8) {
		var key = ad + "|" + strings.ToUpper(ua)
		rdb.Set(key, "1")
		adcount++
	})
	rdb.Flush()
	rdb.Close()

	// 广告数量统计数据 dx_stats , Zj_1461016800, 11111
	lib.StatisticsData("dx_stats", fmt.Sprintf("zj_%s", this.Timestamp),
		convert.ToString(adcount), "incr")
}

func (this *ZjPut) Run() {
	this.kf.AddFun(this.wapData)
	this.kf.WriteFile()       //合成数据
	this.kf.UniqFile()        //合并重复行数据
	this.tagDataStats()       //标签统计
	this.filterData()         //过滤数据,生成ad，ua对应广告id
	this.saveAdvertSet()      //保存广告对应轨迹，并统计每个广告对应的数量
	this.saveTraceToPutSys()  //保存轨迹到投放系统
	this.saveTraceToDianxin() //保存轨迹到电信系统
}

func (this *ZjPut) Clean() {
	this.kf.Clean()
}

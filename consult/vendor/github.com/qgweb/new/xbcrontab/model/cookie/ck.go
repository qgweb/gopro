package cookie

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
		Name:   "cookie_put",
		Usage:  "cookie投放数据",
		Action: putRun,
	}
}

func putRun(ctx *cli.Context) {
	jp := NewCookiePut()
	jp.Run()
	jp.Clean()
}

func NewCookiePut() *CookiePut {
	var zj = &CookiePut{}
	zj.kf = dbfactory.NewKVFile(fmt.Sprintf("./%s.txt", convert.ToString(time.Now().Unix())))
	zj.putTags = make(map[string]map[string]int)
	zj.Timestamp = timestamp.GetHourTimestamp(-1)
	zj.initPutAdverts()
	zj.initPutTags("TAGS_5*", "cookie_", "")
	log.Info(zj.putAdverts)
	log.Info(zj.putTags)
	return zj
}

// 浙江投放
type CookiePut struct {
	kf         *dbfactory.KVFile
	putAdverts map[string]int
	putTags    map[string]map[string]int
	Timestamp  string
}

// 初始化需要投放的广告
func (this *CookiePut) initPutAdverts() {
	rdb, err := lib.GetRedisObj()
	if err != nil {
		log.Fatal(err)
	}
	rdb.SelectDb("0")
	this.putAdverts = make(map[string]int)
	for _, v := range strings.Split(lib.GetConfVal("cookie::province_prefix"), ",") {
		alist := rdb.SMembers(v)
		for _, v := range alist {
			this.putAdverts[v] = 1
		}
	}

	rdb.Close()
}

// 初始化投放标签
func (this *CookiePut) initPutTags(tagkey string, prefix1 string, prefix2 string) {
	rdb, err := lib.GetRedisObj()
	if err != nil {
		log.Fatal(err)
	}
	rdb.SelectDb("0")
	for _, key := range rdb.Keys(tagkey) {
		rkey := strings.TrimPrefix(key, strings.TrimSuffix(tagkey, "*") + "_")
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
func (this *CookiePut) coikieData(out chan interface{}, in chan int8) {
	var datacount = 0
	defer func() {
		// 统计数据 jiangsu_put , url_1461016800, 11111
		lib.StatisticsData("dsource_stats", "zj_" + this.Timestamp + "_cookie",
			convert.ToString(datacount), "")
	}()

	fname := "zhejiang_cookie_" + this.Timestamp
	if err := lib.GetFdbData(fname, func(val string) {
		if v := lib.AddPrefix2(val, "cookie_"); v != "" {
			datacount++
			out <- v
		}
	}); err != nil {
		in <- 1
		return
	}
	log.Info("cookieok")
	in <- 1
}

// 标签数据统计
func (this *CookiePut) tagDataStats() {
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
func (this *CookiePut) filterData() {
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
func (this *CookiePut) saveAdvertSet() {
	tname := "advert_tj_zj_cookie_" + this.Timestamp + "_"
	fname := lib.GetConfVal("cookie::data_path") + tname
	this.kf.IDAdUaSet(fname, func(info map[string]int) {
		tm := this.Timestamp
		for k, v := range info {
			aid := strings.TrimPrefix(k, tname)
			// 广告数量统计数据 advert_stats , zj_1461016800_1111, 11111
			lib.StatisticsData("advert_stats", fmt.Sprintf("zj_cookie_%s_%s", tm, aid),
				convert.ToString(v), "incr")
		}
	}, false)
}

// 保存投放轨迹到投放系统
func (this *CookiePut) saveTraceToPutSys() {
	rdb, err := lib.GetPutRedisObj("put_redis_proxy_url")
	if err != nil {
		log.Error("redis连接失败", err)
		return
	}
	go func() {
		for {
			rdb.Receive()
		}
	}()
	//rdb.SelectDb("1")
	adcount := 0
	this.kf.AdUaIdsSet(func(ad string, ua string, aids map[string]int8) {
		key := ad
		if ua != "ua" {
			key = encrypt.DefaultMd5.Encode(ad + "_" + ua)
		}
		for aid, _ := range aids {
			rdb.HSet(key, "advert:" + aid, aid)
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

func (this *CookiePut) Run() {
	this.kf.AddFun(this.coikieData)
	this.kf.WriteFile()       //合成数据
	this.kf.UniqFile()        //合并重复行数据
	this.tagDataStats()       //标签统计
	this.filterData()         //过滤数据,生成ad，ua对应广告id
	this.saveAdvertSet()      //保存广告对应轨迹，并统计每个广告对应的数量
	this.saveTraceToPutSys()  //保存轨迹到投放系统
}

func (this *CookiePut) Clean() {
	this.kf.Clean()
}

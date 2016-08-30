package sh

import (
	"fmt"
	"github.com/bitly/go-simplejson"
	"github.com/codegangsta/cli"
	"github.com/ngaut/log"
	"github.com/qgweb/new/lib/convert"
	"github.com/qgweb/new/lib/dbfactory"
	"github.com/qgweb/new/lib/timestamp"
	"github.com/qgweb/new/xbcrontab/lib"
	"strings"
	"time"
)

func CliPutData() cli.Command {
	return cli.Command{
		Name:   "shanghai_put",
		Usage:  "上海投放数据",
		Action: putRun,
	}
}

func putRun(ctx *cli.Context) {
	jp := NewShPut()
	jp.Run()
	jp.Clean()
}

// 上海投放
type ShPut struct {
	kf         *dbfactory.KVFile
	putAdverts map[string]int
	putTags    map[string]map[string]int
	Timestamp  string
}

func NewShPut() *ShPut {
	var sh = &ShPut{}
	sh.kf = dbfactory.NewKVFile(fmt.Sprintf("./%s.txt", convert.ToString(time.Now().Unix())))
	sh.putTags = make(map[string]map[string]int)
	sh.Timestamp = timestamp.GetDayTimestamp(-1)
	sh.initPutAdverts()
	sh.initPutTags("TAGS_3*", "tb_", "mg_")
	sh.initPutTags("TAGS_5*", "url_", "")
	log.Warn(sh.putAdverts)
	return sh
}

// 判断是否满足出价的广告
func (this *ShPut) filterPriceAdvert(aid string) bool {
	rdb, err := lib.GetRedisObj()
	if err != nil {
		log.Error(err)
		return false
	}
	defer rdb.Close()
	rdb.SelectDb("0")
	js, err := simplejson.NewJson([]byte(rdb.Get("advert_info:" + aid)))
	if err != nil {
		return false
	}
	log.Info(js.Get("price").String())
	if v, ok := js.Get("price").String(); ok == nil && (convert.ToFloat64(v) >= 0.8) {
		log.Info(1)
		return true
	}
	return false
}

// 初始化需要投放的广告
func (this *ShPut) initPutAdverts() {
	rdb, err := lib.GetRedisObj()
	if err != nil {
		log.Fatal(err)
	}
	rdb.SelectDb("0")
	this.putAdverts = make(map[string]int)
	alist := rdb.SMembers(lib.GetConfVal("shanghai::province_prefix"))
	log.Info(alist)
	for _, v := range alist {
		if this.filterPriceAdvert(v) {
			this.putAdverts[v] = 1
		}
	}
	rdb.Close()
}

// 初始化投放标签
func (this *ShPut) initPutTags(tagkey string, prefix1 string, prefix2 string) {
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

// 域名数据获取
func (this *ShPut) domainData(out chan interface{}, in chan int8) {
	var datacount = 0
	defer func() {
		// 统计数据 jiangsu_put , url_1461016800, 11111
		lib.StatisticsData("dsource_stats", "sh_" + this.Timestamp + "_url",
			convert.ToString(datacount), "")
	}()
	fname := "shanghai_url_" + this.Timestamp
	if err := lib.GetFdbData(fname, func(val string) {
		if v := lib.AddPrefix(val, "url_"); v != "" {
			datacount++
			out <- v
		}
	}); err != nil {
		in <- 1
		return
	}
	log.Info("域名ok")
	in <- 1
}

// 标签数据统计
func (this *ShPut) tagDataStats() {
	fname := convert.ToString(time.Now().UnixNano()) + "_"
	this.kf.IDAdUaSet(fname, func(info map[string]int) {
		for k, v := range info {
			tagid := strings.TrimPrefix(k, fname)
			tagids := strings.Split(tagid, "_")
			// 标签统计数据 tags_stats , url_1461016800, 11111
			lib.StatisticsData("tags_stats", fmt.Sprintf("sh_%s_%s_%s", this.Timestamp, tagids[0], tagids[1]),
				convert.ToString(v), "incr")
		}
	}, true)
}

// 过滤数据
func (this *ShPut) filterData() {
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
func (this *ShPut) saveAdvertSet() {
	tname := "advert_tj_sh_" + this.Timestamp + "_"
	fname := lib.GetConfVal("shanghai::data_path") + tname
	this.kf.IDAdUaSet(fname, func(info map[string]int) {
		tm := this.Timestamp
		for k, v := range info {
			aid := strings.TrimPrefix(k, tname)
			// 广告数量统计数据 advert_stats , zj_1461016800_1111, 11111
			lib.StatisticsData("advert_stats", fmt.Sprintf("sh_%s_%s", tm, aid),
				convert.ToString(v), "")
		}
	}, false)
}

// 保存投放轨迹到投放系统
func (this *ShPut) saveTraceToPutSys() {
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
		for aid, _ := range aids {
			rdb.HSet(key, "advert:" + aid, aid)
		}
		rdb.Expire(key, 86400)
		adcount++
	})
	rdb.Flush()
	rdb.Close()
	// 广告数量统计数据 put_stats , Zj_1461016800, 11111
	lib.StatisticsData("put_stats", fmt.Sprintf("sh_%s", this.Timestamp),
		convert.ToString(adcount), "")
}

func (this *ShPut) Run() {
	this.kf.AddFun(this.domainData)
	this.kf.WriteFile()      //合成数据
	this.tagDataStats()      //标签统计
	this.kf.UniqFile()       //合并重复行数据
	this.filterData()        //过滤数据,生成ad，ua对应广告id
	this.saveAdvertSet()     //保存广告对应轨迹，并统计每个广告对应的数量
	this.saveTraceToPutSys() //保存轨迹到投放系统
}

func (this *ShPut) Clean() {
	this.kf.Clean()
}

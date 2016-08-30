package js

import (
	"bufio"
	"fmt"
	"goclass/convert"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/codegangsta/cli"
	"github.com/ngaut/log"
	"github.com/qgweb/gopro/lib/encrypt"
	"github.com/qgweb/new/lib/dbfactory"
	"github.com/qgweb/new/lib/timestamp"
	"github.com/qgweb/new/xbcrontab/lib"
)

func CliPutData() cli.Command {
	return cli.Command{
		Name:   "jiangsu_put",
		Usage:  "江苏投放数据",
		Action: putRun,
	}
}

func putRun(ctx *cli.Context) {
	jp := NewJsPut()
	jp.Run()
	jp.Clean()
}

func NewJsPut() *JsPut {
	var jp = &JsPut{}
	jp.kf = dbfactory.NewKVFile(fmt.Sprintf("./%s.txt", convert.ToString(time.Now().Unix())))
	jp.putTags = make(map[string]map[string]int)
	jp.initArea()
	jp.initPutAdverts()
	jp.initPutTags("TAGS_3*", "tb_", "mg_")
	jp.initPutTags("TAGS_5*", "url_", "")
	return jp
}

type JsPut struct {
	kf         *dbfactory.KVFile
	areamap    map[string]string
	putAdverts map[string]int
	putTags    map[string]map[string]int
}

// 初始化cox对应区域
func (this *JsPut) initArea() {
	this.areamap = make(map[string]string)
	f, err := os.Open(lib.GetConfVal("jiangsu::areapath"))
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

		this.areamap[info[0]] = strings.TrimSpace(info[1])
	}
	log.Info("江苏区域数量", len(this.areamap))
}

// 初始化需要投放的广告
func (this *JsPut) initPutAdverts() {
	rdb, err := lib.GetRedisObj()
	if err != nil {
		log.Fatal(err)
	}
	rdb.SelectDb("0")
	this.putAdverts = make(map[string]int)
	alist := rdb.SMembers(lib.GetConfVal("jiangsu::province_prefix"))
	for _, v := range alist {
		this.putAdverts[v] = 1
	}
	rdb.Close()
}

// 初始化投放标签
func (this *JsPut) initPutTags(tagkey string, prefix1 string, prefix2 string) {
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
func (this *JsPut) domainData(out chan interface{}, in chan int8) {
	var datacount = 0
	defer func() {
		// 统计数据 jiangsu_put , url_1461016800, 11111
		lib.StatisticsData("dsource_stats", "js_" + timestamp.GetHourTimestamp(-1) + "_url",
			convert.ToString(datacount), "")
	}()

	fname := "jiangsu_url_" + timestamp.GetHourTimestamp(-1)
	if err := lib.GetFdbData(fname, func(val string) {
		if v := lib.AddPrefix(val, "url_"); v != "" {
			datacount++
			out <- v
		}
	}); err != nil {
		in <- 1
		return
	}
	in <- 1
}

// 其他杂项数据获取
func (this *JsPut) otherData(out chan interface{}, in chan int8) {
	var datacount = 0
	defer func() {
		// 统计数据 jiangsu_put , other_1461016800, 11111
		lib.StatisticsData("dsource_stats", "js_" + timestamp.GetHourTimestamp(-1) + "_other",
			convert.ToString(datacount), "")
	}()

	fname := "jiangsu_other_" + timestamp.GetHourTimestamp(-1)
	if err := lib.GetFdbData(fname, func(val string) {
		if v := lib.AddPrefix(val, "mg_"); v != "" {
			datacount++
			out <- v
		}
	}); err != nil {
		in <- 1
		return
	}
	in <- 1
}

// 标签数据统计
func (this *JsPut) tagDataStats() {
	fname := convert.ToString(time.Now().UnixNano()) + "_"
	this.kf.IDAdUaSet(fname, func(info map[string]int) {
		for k, v := range info {
			tagid := strings.TrimPrefix(k, fname)
			tagids := strings.Split(tagid, "_")
			// 标签统计数据 tags_stats , url_1461016800, 11111
			lib.StatisticsData("tags_stats", fmt.Sprintf("js_%s_%s_%s", timestamp.GetHourTimestamp(-1), tagids[0], tagids[1]),
				convert.ToString(v), "incr")
		}
	}, true)
}

// 过滤数据
func (this *JsPut) filterData() {
	this.kf.Filter(func(info dbfactory.AdUaAdverts) (string, bool) {
		var advertIds = make(map[string]int)
		for tagid, _ := range info.AId {
			if v, ok := this.putTags[tagid]; ok {
				for aid, _ := range v {
					advertIds[aid] = 1
				}
			}
		}
		var aids = make([]string, 0, len(advertIds))
		for k, _ := range advertIds {
			aids = append(aids, k)
		}
		if len(aids) != 0 {
			return fmt.Sprintf("%s\t%s\t%s", info.Ad, info.UA, strings.Join(aids, ",")), true
		}
		return "", false
	})
}

// 保存广告对应的ad，ua
func (this *JsPut) saveAdvertSet() {
	tname := "advert_tj_js_" + timestamp.GetHourTimestamp(-1) + "_"
	fname := lib.GetConfVal("jiangsu::data_path") + tname
	this.kf.IDAdUaSet(fname, func(info map[string]int) {
		tm := timestamp.GetHourTimestamp(-1)
		for k, v := range info {
			aid := strings.TrimPrefix(k, tname)
			// 广告数量统计数据 advert_stats , js_1461016800_1111, 11111
			lib.StatisticsData("advert_stats", fmt.Sprintf("js_%s_%s", tm, aid),
				convert.ToString(v), "")
		}
	}, false)
}

// 保存投放轨迹到投放系统
func (this *JsPut) saveTraceToPutSys() {
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
			ua = encrypt.DefaultMd5.Encode(encrypt.DefaultBase64.Decode(ua))
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
	// 广告数量统计数据 put_stats , js_1461016800, 11111
	lib.StatisticsData("put_stats", fmt.Sprintf("js_%s", timestamp.GetHourTimestamp(-1)),
		convert.ToString(adcount), "")
}

// 保存投放轨迹到电信ftp
func (this *JsPut) saveTraceToDianxin() {
	var (
		ftp = lib.GetConfVal("jiangsu::ftp_path")
		ppath = lib.GetConfVal("jiangsu::put_path")
		rk = "account.10046.sha1." + time.Now().Add(-time.Hour).Format("200601021504")
		fname = ppath + "/" + rk
		adcount = 0
	)

	f, err := os.Create(fname)
	if err != nil {
		log.Error("创建文件失败", err)
		return
	}
	defer f.Close()

	this.kf.AdSet(func(ad string) {
		if v, ok := this.areamap[ad]; ok {
			f.WriteString(ad + "," + v + "\n")
			adcount++
		}
	})
	cmd := exec.Command(ftp, rk)
	str, err := cmd.Output()
	log.Info(string(str), err)

	// 广告数量统计数据 dx_stats , js_1461016800, 11111
	lib.StatisticsData("dx_stats", fmt.Sprintf("js_%s", timestamp.GetHourTimestamp(-1)),
		convert.ToString(adcount), "")
}

func (this *JsPut) Run() {
	this.kf.AddFun(this.domainData)
	this.kf.AddFun(this.otherData)
	this.kf.WriteFile()       //合成数据
	this.kf.UniqFile()        //合并重复行数据
	this.tagDataStats()       //标签统计
	this.filterData()         //过滤数据,生成ad，ua对应广告id
	this.saveAdvertSet()      //保存广告对应轨迹，并统计每个广告对应的数量
	this.saveTraceToPutSys()  //保存轨迹到投放系统
	this.saveTraceToDianxin() //保存轨迹到电信系统
}

func (this *JsPut) Clean() {
	this.kf.Clean()
}

// 电商，医疗，金融等其他数据
package middle

import (
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/ngaut/log"
	"github.com/qgweb/go-hbase"
	"github.com/qgweb/gopro/lib/convert"
	"github.com/qgweb/new/lib/encrypt"
	"github.com/qgweb/new/lib/mongodb"
	"github.com/qgweb/new/lib/timestamp"
	"github.com/qgweb/new/xcrontab/common"
	"runtime/debug"
	"strings"
)

func CliUserTrack() cli.Command {
	return cli.Command{
		Name:   "zhejiang_user_trace",
		Usage:  "浙江：用户轨迹每小时数据",
		Action: userRun,
	}
}

func newUserTrack() *UserTrack {
	d := &UserTrack{}
	d.hb = common.CommonHbase
	d.fname = "/tmp/" + timestamp.GetTimestamp() + ".txt"
	d.kvf = common.NewKVFile(d.fname)
	d.mg = common.CommonDataMongo
	conf := mongodb.MongodbQueryConf{Db: "data_source", Table: "useraction_put"}
	d.mgb = mongodb.NewMongodbBufferWriter(d.mg, conf)
	conf = mongodb.MongodbQueryConf{Db: "data_source", Table: "useraction_put_big"}
	d.mg_big, _ = d.mg.Get()
	d.mgb_big = mongodb.NewMongodbBufferWriter(d.mg_big, conf)
	d.bigCategoryMap = make(map[string]string)
	d.getBigCat()
	return d
}

func userRun(c *cli.Context) {
	d := newUserTrack()
	d.run()
	d.clean()
}

type UserTrack struct {
	hb             hbase.HBaseClient
	kvf            *common.KVFile
	mgb            *mongodb.MongodbBufferWriter
	mgb_big        *mongodb.MongodbBufferWriter
	mg             *mongodb.Mongodb
	mg_big         *mongodb.Mongodb
	fname          string
	bigCategoryMap map[string]string
}

func (this *UserTrack) getBigCat() {
	conf := mongodb.MongodbQueryConf{}
	conf.Db = "data_source"
	conf.Table = "taocat"
	conf.Query = mongodb.MM{"type": "0"}
	conf.Select = mongodb.MM{"bid": 1, "cid": 1}
	this.mg.Query(conf, func(info map[string]interface{}) {
		this.bigCategoryMap[convert.ToString(info["cid"])] = convert.ToString(info["bid"])
	})
}

func (this *UserTrack) cleanPutTable() {
	qconf := mongodb.MongodbQueryConf{}
	qconf.Db = "data_source"
	qconf.Table = "useraction_put"
	qconf.Index = []string{"tag.tagId"}
	this.mg.Drop(qconf)
	this.mg.Create(qconf)
	this.mg.EnsureIndex(qconf)
	qconf.Index = []string{"adua"}
	this.mg.EnsureIndex(qconf)
	qconf.Index = []string{"AD"}
	this.mg.EnsureIndex(qconf)

	qconf = mongodb.MongodbQueryConf{}
	qconf.Db = "data_source"
	qconf.Table = "useraction_put_big"
	qconf.Index = []string{"tag.tagId"}
	this.mg_big.Drop(qconf)
	this.mg_big.Create(qconf)
	this.mg_big.EnsureIndex(qconf)
	qconf.Index = []string{"adua"}
	this.mg_big.EnsureIndex(qconf)
	qconf.Index = []string{"AD"}
	this.mg_big.EnsureIndex(qconf)
}

func (this *UserTrack) saveData(info []string) {
	if len(info) < 3 {
		return
	}

	ad := info[0]
	ua := info[1]
	adua := encrypt.DefaultMd5.Encode(ad + ua)
	cids := make([]mongodb.MM, 0)
	cids_put := make([]mongodb.MM, 0)
	for _, v := range strings.Split(info[2], ",") {
		im := 0
		if mongodb.IsObjectId(v) {
			im = 1
		}
		cids = append(cids, mongodb.MM{"tagId": v, "tagmongo": im})

		if bt, ok := this.bigCategoryMap[v]; ok {
			v = bt
		}
		cids_put = append(cids, mongodb.MM{"tagId": v, "tagmongo": im})
	}
	this.mgb.Write(mongodb.MM{
		"AD":   ad,
		"UA":   ua,
		"adua": adua,
		"tag":  cids,
	}, 10000)
	this.mgb_big.Write(mongodb.MM{
		"AD":   ad,
		"UA":   ua,
		"adua": adua,
		"tag":  cids_put,
	}, 10000)
}

func (this *UserTrack) businessData(out chan interface{}, in chan int8) {
	var (
		eghour = timestamp.GetHourTimestamp(-1)
		bghour = timestamp.GetHourTimestamp(-2)
	)
	conf := mongodb.MongodbQueryConf{}
	conf.Db = "data_source"
	conf.Table = "useraction"
	conf.Query = mongodb.MM{"timestamp": mongodb.MM{"$gte": bghour, "$lte": eghour},
		"domainId": mongodb.MM{"$ne": "0"}}
	this.mg.Query(conf, func(info map[string]interface{}) {
		ua := "ua"
		ad := convert.ToString(info["AD"])
		if u, ok := info["UA"]; ok {
			ua = convert.ToString(u)
		}
		cids := make([]string, 0, len(info["tag"].([]interface{})))
		for _, v := range info["tag"].([]interface{}) {
			if tags, ok := v.(map[string]interface{}); ok {
				if strings.TrimSpace(convert.ToString(tags["tagId"])) != "" {
					cids = append(cids, convert.ToString(tags["tagId"]))
				}
			}
		}
		out <- fmt.Sprintf("%s\t%s\t%s", ad, ua, strings.Join(cids, ","))

	})
	in <- 1
}

func (this *UserTrack) otherData(out chan interface{}, in chan int8) {
	var (
		eghour = timestamp.GetHourTimestamp(-1)
		bghour = timestamp.GetHourTimestamp(-25)
	)
	conf := mongodb.MongodbQueryConf{}
	conf.Db = "data_source"
	conf.Table = "useraction"
	conf.Query = mongodb.MM{"timestamp": mongodb.MM{"$gte": bghour, "$lte": eghour}, "domainId": "0"}
	this.mg_big.Query(conf, func(info map[string]interface{}) {
		ua := "ua"
		ad := convert.ToString(info["AD"])
		if u, ok := info["UA"]; ok {
			ua = convert.ToString(u)
		}
		cids := make([]string, 0, len(info["tag"].([]interface{})))
		for _, v := range info["tag"].([]interface{}) {
			if tags, ok := v.(map[string]interface{}); ok {
				if strings.TrimSpace(convert.ToString(tags["tagId"])) != "" {
					cids = append(cids, convert.ToString(tags["tagId"]))
				}
			}
		}
		out <- fmt.Sprintf("%s\t%s\t%s", ad, ua, strings.Join(cids, ","))
	})
	in <- 1
}

func (this *UserTrack) run() {
	defer func() {
		if err := recover(); err != nil {
			log.Error(err)
			debug.PrintStack()
		}
	}()
	this.kvf.AddFun(this.businessData)
	this.kvf.AddFun(this.otherData)
	log.Error(this.kvf.WriteFile())
	this.cleanPutTable()
	log.Error(this.kvf.Origin(this.saveData))
	this.mgb.Flush()
	this.mgb_big.Flush()
}

func (this *UserTrack) clean() {
	this.kvf.Clean()
	this.mg.Close()
	this.mg_big.Close()
}

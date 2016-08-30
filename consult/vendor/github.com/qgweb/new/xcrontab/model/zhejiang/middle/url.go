// 跑域名中间数据
package middle

import (
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/ngaut/log"
	"github.com/qgweb/go-hbase"
	"github.com/qgweb/new/lib/encrypt"
	"github.com/qgweb/new/lib/mongodb"
	"github.com/qgweb/new/lib/timestamp"
	"github.com/qgweb/new/xcrontab/common"
	"runtime/debug"
	"strings"
	"time"
)

func CliDomain() cli.Command {
	return cli.Command{
		Name:   "zhejiang_urltrack",
		Usage:  "浙江：域名每小时数据",
		Action: domainRun,
	}
}

func newDomain() *Domain {
	d := &Domain{}
	d.hb = common.CommonHbase
	d.fname = "/tmp/" + timestamp.GetTimestamp() + ".txt"
	d.kvf = common.NewKVFile(d.fname)
	d.mg = common.CommonDataMongo
	conf := mongodb.MongodbQueryConf{Db: "data_source", Table: "urltrack_put"}
	d.mgb = mongodb.NewMongodbBufferWriter(d.mg, conf)

	return d
}

func domainRun(c *cli.Context) {
	d := newDomain()
	d.run()
	d.clean()
}

type Domain struct {
	hb    hbase.HBaseClient
	kvf   *common.KVFile
	mgb   *mongodb.MongodbBufferWriter
	mg    *mongodb.Mongodb
	fname string
}

func (this *Domain) cleanPutTable() {
	qconf := mongodb.MongodbQueryConf{}
	qconf.Db = "data_source"
	qconf.Table = "urltrack_put"
	qconf.Index = []string{"cids.id"}
	this.mg.Drop(qconf)
	this.mg.Create(qconf)
	this.mg.EnsureIndex(qconf)
	qconf.Index = []string{"adua"}
	this.mg.EnsureIndex(qconf)
}

func (this *Domain) domainData(out chan interface{}, in chan int8) {
	var (
		tname = "zhejiang_urltrack_" + time.Now().Format("200601")
		btime = timestamp.GetHourTimestamp(-1)
		etime = timestamp.GetHourTimestamp(0)
	)

	sc := hbase.NewScan([]byte(tname), 10000, this.hb)
	sc.StartRow = []byte(btime)
	sc.StopRow = []byte(etime)

	for {
		row := sc.Next()
		if row == nil {
			break
		}

		ad := strings.TrimSpace(string(row.Columns["base:ad"].Value))
		ua := strings.TrimSpace(string(row.Columns["base:ua"].Value))
		cids := make([]string, 0, len(row.Columns)-2)

		for _, v := range row.Columns {
			if string(v.Family) == "cids" {
				cids = append(cids, string(v.Qual))
			}
		}

		out <- fmt.Sprintf("%s\t%s\t%s", ad, ua, strings.Join(cids, ","))
	}
	in <- 1
	sc.Close()
}

func (this *Domain) saveData(info []string) {
	if len(info) < 3 {
		return
	}

	ad := info[0]
	ua := info[1]
	adua := encrypt.DefaultMd5.Encode(ad + ua)
	cids := make([]mongodb.MM, 0)
	for _, v := range strings.Split(info[2], ",") {
		cids = append(cids, mongodb.MM{"id": v})
	}
	this.mgb.Write(mongodb.MM{
		"ad":   ad,
		"ua":   ua,
		"adua": adua,
		"cids": cids,
	}, 10000)
}

func (this *Domain) run() {
	defer func() {
		if err := recover(); err != nil {
			log.Error(err)
			debug.PrintStack()
		}
	}()

	this.kvf.AddFun(this.domainData)
	log.Error(this.kvf.WriteFile())
	this.cleanPutTable()
	log.Error(this.kvf.Origin(this.saveData))
	this.mgb.Flush()
}

func (this *Domain) checkError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func (this *Domain) clean() {
	this.kvf.Clean()
	this.hb.Close()
	this.mg.Close()
}

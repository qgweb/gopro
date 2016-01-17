package domain

import (
	"github.com/codegangsta/cli"
	"github.com/juju/errors"
	"github.com/ngaut/log"
	"github.com/qgweb/go-hbase"
	"github.com/qgweb/gopro/9xutool/common"
	"github.com/qgweb/new/lib/config"
	"github.com/qgweb/new/lib/encrypt"
	"github.com/qgweb/new/lib/mongodb"
	"github.com/qgweb/new/lib/rediscache"
	"github.com/qgweb/new/lib/timestamp"
	"runtime/debug"
)

type UrlTrace struct {
	mdb       *mongodb.Mongodb
	mdbWriter *mongodb.MongodbBufferWriter
	conf      *config.Config
	robj      *rediscache.MemCache
	hobj      hbase.HBaseClient
	keyprefix string
}

func newUrltrace() (ut *UrlTrace, err error) {
	ut = &UrlTrace{}
	iname := common.GetBasePath() + "/conf/ut.conf"
	if ut.conf, err = common.GetConfObj(iname); err != nil {
		return nil, errors.Trace(err)
	}
	if ut.mdb, err = common.GetMongoObj(iname, "mongo"); err != nil {
		return nil, errors.Trace(err)
	}
	if ut.robj, err = common.GetRedisObj(iname, "redis_cache"); err != nil {
		return nil, errors.Trace(err)
	}
	if ut.hobj, err = common.GetHbaseObj(iname, "hbase"); err != nil {
		return nil, errors.Trace(err)

	}
	ut.mdbWriter = ut.getWriter()
	ut.keyprefix = mongodb.GetObjectId() + "_"
	return
}

func NewURLTraceCli() cli.Command {
	return cli.Command{
		Name:  "url_trace_merge",
		Usage: "生成域名1小时的数据,供九旭精准投放",
		Action: func(c *cli.Context) {
			defer func() {
				if msg := recover(); msg != nil {
					log.Error(msg)
					debug.PrintStack()
				}
			}()

			ut, err := newUrltrace()
			if err != nil {
				log.Info(errors.Details(err))
				return
			}
			ut.Do()
			ut.Clean()
		},
	}
}

func (this *UrlTrace) getWriter() *mongodb.MongodbBufferWriter {
	qconf := mongodb.MongodbQueryConf{}
	qconf.Db = "data_source"
	qconf.Table = "urltrack_put"
	return mongodb.NewMongodbBufferWriter(this.mdb, qconf)
}

func (this *UrlTrace) Clean() {
	this.mdb.Close()
	this.robj.Clean(this.keyprefix)
	this.hobj.Close()
}

func (this *UrlTrace) Read(saveFun func(hbase.ResultRow)) {
	var (
		tableName = "zhejiang_urltrace_201601"
		beginTime = timestamp.GetHourTimestamp(-2) + "_"
		endTime   = timestamp.GetHourTimestamp(0) + "_"
	)
	scn := hbase.NewScan([]byte(tableName), 10000, this.hobj)
	scn.StartRow = []byte(beginTime)
	scn.StopRow = []byte(endTime)

	for {
		row := scn.Next()
		if row == nil {
			break
		}

		saveFun(row)
	}
	scn.Close()
}

func (this *UrlTrace) SaveOne(row *hbase.ResultRow) {
	ad := string(row.Columns["base:ad"].Value)
	ua := string(row.Columns["base:ua"].Value)
	cids := make([]mongodb.MM, 0, len(row.Columns)-2)

	for _, v := range row.Columns {
		if v.Family == []byte("cids") {
			cids = append(cids, mongodb.MM{"id", string(v.Value)})
		}
	}

	info := mongodb.MM{
		"ad":   ad,
		"ua":   ua,
		"cids": cids,
		"adua": encrypt.DefaultMd5.Encode(ad + ua)}
	this.mdbWriter.Write(info, 10000)
}

func (this *UrlTrace) FlushWriter() {
	this.mdbWriter.Flush()
}

func (this *UrlTrace) Do() {
	this.Read(this.SaveOne)
	this.FlushWriter()
}

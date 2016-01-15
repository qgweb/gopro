// 访客找回
package visitor

import (
	"github.com/codegangsta/cli"
	"github.com/juju/errors"
	"github.com/ngaut/log"
	"github.com/qgweb/go-hbase"
	"github.com/qgweb/new/lib/common"
	"github.com/qgweb/new/lib/config"
	"github.com/qgweb/new/lib/convert"
	"github.com/qgweb/new/lib/encrypt"
	"github.com/qgweb/new/lib/mongodb"
	"github.com/qgweb/new/lib/rediscache"
	"github.com/qgweb/new/lib/set"
	"github.com/qgweb/new/lib/timestamp"
	"runtime/debug"
	"strings"
)

type ZhejiangVisitor struct {
	iniFile    config.ConfigContainer
	mg         *mongodb.Mongodb
	mgcookie   *mongodb.Mongodb
	rsCache    *rediscache.MemCache
	rs         *rediscache.MemCache
	hb         hbase.HBaseClient
	keyprefix  string
	packageKey string
}

type visitorParam struct {
	advertId string
	pkgId    string
	expire   int
}

func NewZhejiangVisitor() (*ZhejiangVisitor, error) {
	var zv = &ZhejiangVisitor{}
	var err error

	zv.keyprefix = mongodb.GetObjectId() + "_"
	zv.packageKey = "PERSON_PACKAGE"
	zv.iniFile, err = config.NewConfig("ini", common.GetBasePath()+"/conf/"+"visitor.ini")
	if err != nil {
		return nil, errors.Trace(err)
	}

	mf := mongodb.MongodbConf{}
	mf.Host = zv.iniFile.String("mongo::host")
	mf.Port = zv.iniFile.String("mongo::port")
	mf.Db = zv.iniFile.String("mongo::db")
	mf.UName = zv.iniFile.String("mongo::user")
	mf.Upwd = zv.iniFile.String("mongo::pwd")
	zv.mg, err = mongodb.NewMongodb(mf)
	if err != nil {
		return nil, errors.Trace(err)
	}

	mf.Host = zv.iniFile.String("mongo-cookie::host")
	mf.Port = zv.iniFile.String("mongo-cookie::port")
	mf.Db = zv.iniFile.String("mongo-cookie::db")
	mf.UName = zv.iniFile.String("mongo-cookie::user")
	mf.Upwd = zv.iniFile.String("mongo-cookie::pwd")
	zv.mgcookie, err = mongodb.NewMongodb(mf)
	if err != nil {
		return nil, errors.Trace(err)
	}

	rf := rediscache.MemConfig{}
	rf.Host = zv.iniFile.String("redis-cache::host")
	rf.Port = zv.iniFile.String("redis-cache::port")
	zv.rsCache, err = rediscache.New(rf)
	if err != nil {
		return nil, errors.Trace(err)
	}
	zv.rsCache.SelectDb(zv.iniFile.String("redis-cache::db"))

	rf = rediscache.MemConfig{}
	rf.Host = zv.iniFile.String("redis::host")
	rf.Port = zv.iniFile.String("redis::port")
	zv.rs, err = rediscache.New(rf)
	if err != nil {
		return nil, errors.Trace(err)
	}
	zv.rs.SelectDb(zv.iniFile.String("redise::db"))

	hosts := strings.Split(zv.iniFile.String("hbase::host"), ",")
	for k, _ := range hosts {
		hosts[k] += ":" + zv.iniFile.String("hbase::port")
	}

	zv.hb, err = hbase.NewClient(hosts, "/hbase")
	if err != nil {
		return nil, errors.Trace(err)
	}

	return zv, nil
}

func NewZhejiangVisitorCli() cli.Command {
	return cli.Command{
		Name:  "package_visitor",
		Usage: "生成浙江访客找回的数据",
		Action: func(c *cli.Context) {
			defer func() {
				if msg := recover(); msg != nil {
					log.Error(msg)
					debug.PrintStack()
				}
			}()

			zv, err := NewZhejiangVisitor()
			if err != nil {
				log.Error(errors.Details(err))
				return
			}

			zv.Do()
			zv.Clean()
		}}
}

func (this *ZhejiangVisitor) Clean() {
	this.rsCache.Clean(this.keyprefix)
	this.mg.Close()
	this.rsCache.Close()
	this.iniFile = nil
}

// 获取人群找回包信息
func (this *ZhejiangVisitor) visitorPackage() []visitorParam {
	var list = make([]visitorParam, 0)
	for _, v := range this.rs.SMembers(this.packageKey) {
		var vp visitorParam
		infos := strings.Split(v, "_")
		if len(infos) != 3 {
			continue
		}
		vp.pkgId = infos[0]
		vp.expire = convert.ToInt(infos[1])
		vp.advertId = infos[2]
		list = append(list, vp)
	}
	return list
}

func (this *ZhejiangVisitor) visitorCookies(vps []visitorParam) map[string]*set.Set {
	list := make(map[string]*set.Set)
	for _, vp := range vps {
		cookies := set.New()
		sc := hbase.NewScan([]byte("domain-cookie"), 10000, this.hb)
		sc.StartRow = []byte(vp.pkgId + "_" + timestamp.GetDayTimestamp(vp.expire*-1))
		sc.StopRow = []byte(vp.pkgId + "_" + timestamp.GetDayTimestamp(0))
		sc.AddStringColumn("base", "cookie")

		for {
			rw := sc.Next()
			if rw == nil {
				break
			}
			cookies.Add(string(rw.Columns["base:cookie"].Value))
		}
		if cookies.Len() > 0 {
			list[vp.advertId] = cookies
		}
		sc.Close()
	}
	return list
}

func (this *ZhejiangVisitor) getAdUaByMongo(cookie string) (string, error) {
	qconf := mongodb.MongodbQueryConf{}
	qconf.Db = this.mgcookie.GetConf().Db
	qconf.Table = "dt_user"
	qconf.Query = mongodb.MM{"_id": mongodb.ObjectId(cookie)}
	qconf.Select = mongodb.MM{"cox": 1, "ua": 1}
	info, err := this.mgcookie.One(qconf)
	if err != nil {
		return "", errors.Trace(err)
	}
	if convert.ToString(info["cox"]) == "" {
		return "", errors.New("cox为空")
	}
	return convert.ToString(info["cox"]) + "_" +
		encrypt.DefaultBase64.Encode(convert.ToString(info["ua"])), nil
}

func (this *ZhejiangVisitor) getAdUaByHbase(cookie string) (string, error) {
	get := hbase.NewGet([]byte(cookie))
	get.AddStringColumn("base", "cox")
	get.AddStringColumn("base", "ua")
	row, err := this.hb.Get("xu-cookie", get)
	if err != nil {
		return "", errors.Trace(err)
	}
	if string(row.Columns["base:cox"].Value) == "" {
		return "", errors.New("cox为空")
	}
	return convert.ToString(row.Columns["base:cox"].Value) + "_" +
		encrypt.DefaultBase64.Encode(convert.ToString(row.Columns["base:ua"].Value)), nil
}

func (this *ZhejiangVisitor) visitorAdUa(vcs map[string]*set.Set) {
	for k, v := range vcs {
		for _, cookie := range v.List() {
			adua, err := this.getAdUaByMongo(convert.ToString(cookie))
			if err != nil {
				adua, err = this.getAdUaByHbase(convert.ToString(cookie))
				if err != nil {
					continue
				}
				this.rsCache.HSet(this.keyprefix+adua, k, "1")
				continue
			}
			this.rsCache.HSet(this.keyprefix+adua, k, "1")
		}
		this.rsCache.Flush()
	}
}

func (this *ZhejiangVisitor) Save() {
	var conf = mongodb.MongodbQueryConf{}
	conf.Db = this.mg.GetConf().Db
	conf.Table = "zhejiang_visitor"
	conf.Insert = make([]interface{}, 0)
	mbw := mongodb.NewMongodbBufferWriter(this.mg, conf)
	this.mg.Drop(conf)
	this.mg.Create(conf)
	for _, key := range this.rsCache.Keys(this.keyprefix + "*") {
		aduas := strings.Split(strings.TrimPrefix(key, this.keyprefix), "_")
		adids := this.rsCache.HGetAllKeys(key)
		mbw.Write(mongodb.MM{"ad": aduas[0], "ua": aduas[1], "aids": adids}, 10000)
	}
	mbw.Flush()
}

func (this *ZhejiangVisitor) Debug(vv interface{}) {
	if v := this.iniFile.String("default::debug"); v == "1" {
		log.Debug(vv)
	}
}

func (this *ZhejiangVisitor) Do() {
	var a = this.visitorPackage()
	this.Debug(a)
	var b = this.visitorCookies(a)
	this.Debug(b)
	this.visitorAdUa(b)
	this.Save()
}

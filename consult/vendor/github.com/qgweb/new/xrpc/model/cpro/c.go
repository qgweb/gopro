package cpro

import (
	"github.com/bradfitz/gomemcache/memcache"
	"github.com/juju/errors"
	"github.com/qgweb/go-hbase"
	"github.com/qgweb/new/lib/encrypt"
	"github.com/qgweb/new/lib/timestamp"
	"github.com/qgweb/new/xrpc/db"
	"gopkg.in/mgo.v2/bson"
)

type CproData struct {
}

// cookie参数
type CookieParam struct {
	id     string
	cox    string
	ua     string
	date   string
	ip     string
	pid    string
	cid    string
	is_new string
}

// 投放广告记录参数
type AdvertPutParam struct {
	ua     string //user-agent
	ad     string //cox
	pv     string //pv
	click  string //点击
	advert string //广告id
}

// 解析cookie参数
func parseCookieParam(param map[string]string) (cp CookieParam) {
	if v, ok := param["id"]; ok {
		cp.id = v
	}
	if v, ok := param["cox"]; ok {
		cp.cox = v
	}
	if v, ok := param["ua"]; ok {
		cp.ua = v
	}
	if v, ok := param["date"]; ok {
		cp.date = v
	}
	if v, ok := param["ip"]; ok {
		cp.ip = v
	}
	if v, ok := param["pid"]; ok {
		cp.pid = v
	}
	if v, ok := param["cid"]; ok {
		cp.cid = v
	}
	if v, ok := param["is_new"]; ok {
		cp.is_new = v
	}
	return
}

// 解析广告参数
func parseAdvertParam(param map[string]string) (ap AdvertPutParam) {
	if v, ok := param["advert"]; ok {
		ap.advert = v
	}
	if v, ok := param["ad"]; ok {
		ap.ad = v
	}
	if v, ok := param["ua"]; ok {
		ap.ua = v
	}
	if v, ok := param["pv"]; ok {
		ap.pv = v
	}
	if v, ok := param["click"]; ok {
		ap.click = v
	}
	return
}

// 创建hbase表
func (this CproData) createTable(conn hbase.HBaseClient, tableName string) error {
	ok, err := conn.TableExists(tableName)
	if err != nil {
		return err
	}

	if !ok {
		td := hbase.NewTableDesciptor(tableName)
		td.AddColumnDesc(hbase.NewColumnFamilyDescriptor("base"))
		if err := conn.CreateTable(td, nil); err != nil {
			return err
		}
	}
	return nil
}

// 记录cookie
func (this CproData) ReocrdCookie(param map[string]string) error {
	var (
		cp        = parseCookieParam(param)
		conn      = db.GetHbaseConn()
		tableName = "xu-cookie"
	)

	if err := this.createTable(conn, tableName); err != nil {
		return err
	}

	if !bson.IsObjectIdHex(cp.id) {
		return errors.New("cookie-id参数错误")
	}

	put := hbase.NewPut([]byte(cp.id))
	put.AddStringValue("base", "cox", cp.cox)
	put.AddStringValue("base", "ua", cp.ua)
	put.AddStringValue("base", "date", cp.date)
	put.AddStringValue("base", "ip", cp.ip)
	put.AddStringValue("base", "pid", cp.pid)
	put.AddStringValue("base", "cid", cp.cid)
	put.AddStringValue("base", "is_new", cp.is_new)
	conn.Put(tableName, put)
	return nil
}

func (this CproData) ReocrdCookieEx(param map[string]string) error {
	var (
		cp        = parseCookieParam(param)
		conn      = db.GetHbaseConn()
		tableName = "xu-cookie-ex"
	)

	if err := this.createTable(conn, tableName); err != nil {
		return err
	}

	if !bson.IsObjectIdHex(cp.id) {
		return errors.New("cookie-id参数错误")
	}

	put := hbase.NewPut([]byte(bson.NewObjectId().Hex()))
	put.AddStringValue("base", "ckid", cp.id)
	put.AddStringValue("base", "cox", cp.cox)
	put.AddStringValue("base", "ua", cp.ua)
	put.AddStringValue("base", "date", cp.date)
	put.AddStringValue("base", "ip", cp.ip)
	put.AddStringValue("base", "pid", cp.pid)
	put.AddStringValue("base", "cid", cp.cid)
	put.AddStringValue("base", "is_new", cp.is_new)
	conn.Put(tableName, put)
	return nil
}

// 域名访客找回
func (this CproData) DomainVisitor(pkgId string, cookie string, domain string) error {
	var (
		conn      = db.GetHbaseConn()
		date      = timestamp.GetDayTimestamp(0)
		tableName = "domain-cookie"
	)

	if err := this.createTable(conn, tableName); err != nil {
		return err
	}

	put := hbase.NewPut([]byte(pkgId + "_" + date + "_" + cookie))
	put.AddStringValue("base", "date", timestamp.GetDayTimestamp(0))
	put.AddStringValue("base", "cookie", cookie)
	put.AddStringValue("base", "domain", domain)
	conn.Put(tableName, put)
	return nil
}

// 域名生效
func (this CproData) DomainEffect(id string) error {
	var (
		mem  = db.GetMemcacheConn()
		msql = db.GetMysqlConn()
		key  = "DOMAIN_COOKIE_" + id
	)

	it, err := mem.Get(key)
	if err != nil && err != memcache.ErrCacheMiss {
		return err
	}

	if it == nil {
		r, err := msql.Raw("update nxu_group_pkg set is_effective=?,effective_time=? where id=?", 1,
			timestamp.GetTimestamp(), id).Exec()
		if err != nil {
			return err
		}
		if n, err := r.RowsAffected(); err == nil && n > 0 {
			mem.Set(&memcache.Item{Key: key, Value: []byte("1")})
		}
	}

	return nil
}

// 记录广告投放信息
// 包括点击，pv,ad,ua等
func (this CproData) RecordAdvertPutInfo(param map[string]string) error {
	var (
		ap        = parseAdvertParam(param)
		conn      = db.GetHbaseConn()
		tableName = "advert-put-record"
		key       = ap.advert + "_" + encrypt.DefaultMd5.Encode(ap.ad+encrypt.DefaultBase64.Encode(ap.ua))
		merr      string
	)

	if err := this.createTable(conn, tableName); err != nil {
		return err
	}

	put := hbase.NewPut([]byte(key))
	put.AddStringValue("base", "ad", ap.ad)
	put.AddStringValue("base", "ua", encrypt.DefaultBase64.Encode(ap.ua))
	put.AddStringValue("base", "aid", ap.advert)

	incr := hbase.NewIncr([]byte(key))
	if ap.pv == "1" {
		incr.AddStringValue("base", "pv", 1)
	}
	if ap.click == "1" {
		incr.AddStringValue("base", "click", 1)
	}

	if _, err := conn.Put(tableName, put); err != nil {
		merr = merr + err.Error()
	}
	if _, err := conn.Incr(tableName, incr); err != nil {
		merr = merr + "|" + err.Error()
	}

	if merr != "" {
		return errors.New(merr)
	}

	return nil
}

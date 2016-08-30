package main

import (
	"bytes"
	"flag"
	"os/exec"

	"github.com/ngaut/log"
	"github.com/qgweb/new/lib/encrypt"
	"gopkg.in/mgo.v2"
	"time"
	"github.com/nsqio/go-nsq"
	"strings"
	"fmt"
	"github.com/wangtuanjie/ip17mon"
	"sync"
	"github.com/qgweb/gopro/lib/orm"
	"net/url"
	"errors"
	"github.com/qgweb/new/lib/convert"
	"github.com/qgweb/new/lib/timestamp"
	"encoding/json"
)

var (
	nsqhost = flag.String("nsqhost", "127.0.0.1:4150", "nsq连接地址")
	mdbhost = flag.String("mdbhost", "192.168.1.199:27017", "mongodb连接地址")
	dbhost = flag.String("mysqlhost", "root:123456@tcp(192.168.1.199:3306)/ssp?charset=utf8", "mysql地址")
	scmd = flag.String("cmd", "php", "命令")
	sfile = flag.String("cfile", "", "执行的文件")
	mdbSess *mgo.Session
	dbConn *orm.QGORM
)

func init() {
	flag.Parse()
	if *scmd == "" || *sfile == "" {
		log.Fatal("命令参数不能为空")
	}
	initMgodb()
	initIp()
	initMySql()

}

func initMgodb() {
	sess, err := mgo.DialWithTimeout("mongodb://" + *mdbhost, time.Second * 30)
	if err != nil {
		log.Fatal(err)
	}
	mdbSess = sess
}

func initMySql() {
	dbConn = orm.NewORM()
	dbConn.Open(*dbhost)
	dbConn.SetMaxIdleConns(5)
	dbConn.SetMaxOpenConns(5)
}

func initIp() {
	if err := ip17mon.Init("./17monipdb.dat"); err != nil {
		log.Fatal(err)
	}
}

type MLog struct {
	Province string `bson:"province"`
	City     string `bson:"city"`
	AdwId    int `bson:"adw_id"`
	AdId     int `bson:"ad_id"`
	PV       int `bson:"pv"`
	Click    int `bson:"click"`
	Money    float32 `bson:"money"`
	Source   int `bson:"source"`     //来源(0 站内  1站外)
	Domain   string `bson:"domain"`
	Url      string `bson:"url"`
	AD       string `bson:"ad"`
	UA       string `bson:"ua"`
	Cookie   string `bson:"cookie"`
	CusId    string `bson:"string"`  //客户id（ad+ua+cookie）
	Clock    int `bson:"clock"`
	Date     int `bson:"date"`
	Fmoney   float32 `bson:"fmoney"` //折扣后的钱
}

type TailHandler struct {
	lock     sync.Mutex
	SouceMap map[string]byte
}

func (th *TailHandler) Script(buf []byte) []byte {
	cmd := exec.Command(*scmd, *sfile, encrypt.DefaultBase64.Encode(string(buf)))
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Start()
	if err != nil {
		return nil
	}
	err = cmd.Wait()
	if err != nil {
		return nil
	}
	return out.Bytes()
}

func (th *TailHandler) pv_param(info string) (url.Values, error) {
	infos := strings.Split(info, " ")
	if len(infos) < 0 {
		return nil, errors.New("参数格式错误")
	}
	u, err := url.ParseQuery(infos[1][strings.Index(infos[1], "?") + 1:])
	if _, ok := u["lftu"]; !ok {
		u["lftu"] = []string{""}
	}
	return u, err
}

func (th *TailHandler) get_ad(info string) string {
	v, err := url.QueryUnescape(info)
	if err != nil {
		return ""
	}
	ul, err := url.Parse(v)
	if err != nil {
		return ""
	}
	if c := ul.Query().Get("cox"); c != "" {
		return c
	}
	if c := ul.Query().Get("js_cox"); c != "" {
		return c
	}
	if c := ul.Query().Get("sh_cox"); c != "" {
		return c
	}
	return ""
}

func (th *TailHandler) get_url(ltu string, ftu string) string {
	if ftu == "" {
		v, err := url.QueryUnescape(ltu)
		if err != nil {
			return ""
		}
		return v
	}
	return ftu
}

func (th *TailHandler) get_domain(uurl string) string {
	if u, err := url.Parse(uurl); err == nil {
		return u.Scheme + "://" + u.Host
	}
	return ""
}

func (th *TailHandler) get_cookie(info string) string {
	var mm = make(map[string]string)
	for _, v := range strings.Split(info, ";") {
		vv := strings.Split(strings.TrimSpace(v), "=")
		mm[vv[0]] = vv[1]
	}

	if v, ok := mm["dt_uid"]; ok {
		return v
	}
	return ""
}
func (th *TailHandler) Pv(info []string) {
	var ml MLog
	ud, err := th.pv_param(info[3])
	if err != nil {
		return
	}
	p, c := th.getIp(info)
	ml.Province = p
	ml.City = c
	ml.AdwId = convert.ToInt(ud["pd"][0])
	ml.AdId = convert.ToInt(ud["hd"][0])
	ml.UA = info[7]
	ml.AD = th.get_ad(ud["ltu"][0])
	ml.Click = 0
	ml.Clock = convert.ToInt(time.Now().Format("15"))
	ml.Date = convert.ToInt(timestamp.GetDayTimestamp(0))
	ml.Url = th.get_url(ud["ltu"][0], ud["lftu"][0])
	ml.Domain = th.get_domain(ml.Url)
	ml.Cookie = th.get_cookie(info[9])
	ml.CusId = encrypt.DefaultMd5.Encode(ml.AD + ml.UA + ml.Cookie)
	ml.Fmoney = 0
	ml.Money = 0
	ml.PV = 1
	if _, ok := th.SouceMap[ud["pd"][0]]; ok {
		ml.Source = 1
	}
	//fmt.Println(ml)
	th.SaveData(ml)
}

func (th *TailHandler) getmap(info map[string]string, str string) string {
	if v, ok := info[str]; ok {
		return v
	}
	return ""
}

func (th *TailHandler) click_money(adwid int, aid int) (f1 float32, f2 float32) {
	var rate float32 = 1
	sql := dbConn.BSQL().Select("rate").From("nxu_advertising").Where("id=?").GetSQL()
	info, err := dbConn.Get(sql, adwid)
	if err != nil {
		rate = 1
	}
	rate = convert.ToFloat32(info["rate"])
	//出价
	sql = dbConn.BSQL().Select("price").From("nxu_advert").Where("id=?").GetSQL()
	info, err = dbConn.Get(sql, aid)
	if err != nil {
		return 0, 0
	}
	f1 = convert.ToFloat32(info["price"])
	return f1, convert.ToFloat32(fmt.Sprintf("%.2f", f1 * rate))
}

func (th *TailHandler) Click(info []string) {
	var exstr = th.Script([]byte(strings.Join(info, "\t")))
	var exmap map[string]string
	if err := json.Unmarshal(exstr, &exmap); err != nil {
		return
	}
	if len(exmap) == 0 {
		return
	}
	log.Info(exmap)
	var ml MLog
	p, c := th.getIp(info)
	ml.Province = p
	ml.City = c
	ml.AdwId = convert.ToInt(th.getmap(exmap, "pd"))
	ml.AdId = convert.ToInt(th.getmap(exmap, "hd"))
	ml.UA = th.getmap(exmap, "ua")
	ml.AD = th.getmap(exmap, "ad")
	ml.Click = 1
	ml.Clock = convert.ToInt(time.Now().Format("15"))
	ml.Date = convert.ToInt(timestamp.GetDayTimestamp(0))
	ml.Url = th.get_url(th.getmap(exmap, "ltu"), th.getmap(exmap, "lftu"))
	ml.Domain = th.get_domain(ml.Url)
	ml.Cookie = th.get_cookie(info[9])
	ml.CusId = encrypt.DefaultMd5.Encode(ml.AD + ml.UA + ml.Cookie)
	ml.Money, ml.Fmoney = th.click_money(ml.AdwId, ml.AdId)
	if _, ok := th.SouceMap[th.getmap(exmap, "pd")]; ok {
		ml.Source = 1
	}
	th.SaveData(ml)
}

func (th *TailHandler) SaveData(ml MLog) {
	var tb = "log_" + time.Now().Format("2006-01-02")
	log.Error(mdbSess.DB("log").C(tb).Insert(ml))
}

// 0 $remote_addr
// 1 $remote_user
// 2 [$time_local]
// 3 $request"'
// 4 $status
// 5 $body_bytes_sent
// 6 $http_referer
// 7 $http_user_agent
// 8 $http_x_forwarded_for
// 9 $XuCookie
// 10 $geoip_region
// 11 $geoip_city';

func (th *TailHandler) getIp(info []string) (string, string) {
	if info[8] == "-" || info[8] == "" {
		return th.Ip(info[0])
	}
	return th.Ip(info[8])
}

func (th *TailHandler) Ip(ip string) (string, string) {
	th.lock.Lock()
	defer th.lock.Unlock()
	if loc, err := ip17mon.Find(ip); err == nil {
		return loc.Region, loc.City
	}
	return "", ""
}

func (th *TailHandler) HandleMessage(m *nsq.Message) error {
	info := string(m.Body)
	infos := strings.Split(info, "\t")
	if infos[0] == "pv" {
		th.Pv(infos[1:])
	}
	if infos[0] == "click" {
		th.Click(infos[1:])
	}
	return nil
}

func getSourceMap() map[string]byte {
	sql := dbConn.BSQL().Select("distinct(advertising_id) as id").From("nxu_advertising_order").GetSQL()
	var mm = make(map[string]byte)
	if info, err := dbConn.Query(sql); err == nil {
		for _, v := range info {
			mm[v["id"]] = 1
		}
	}
	return mm
}

func main() {
	cus, err := nsq.NewConsumer("log", "logxx", nsq.NewConfig())
	if err != nil {
		log.Fatal(err)
	}

	cus.AddHandler(&TailHandler{SouceMap:getSourceMap()})

	if err := cus.ConnectToNSQD(*nsqhost); err != nil {
		log.Fatal(err)
	}

	for {
		select {
		case <-cus.StopChan:
			return
		}
	}

}

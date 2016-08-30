// 通过ad获取经纬度
package main

import (
	"flag"
	"fmt"
	"github.com/ngaut/log"
	"github.com/qgweb/new/lib/convert"
	"github.com/qgweb/new/lib/mongodb"
	"net/http"
	"strings"
)

type FlagParam struct {
	httpHost   string
	mongoHost  string
	mongoPort  string
	mongoUname string
	mongoUpwd  string
	mongoDb    string
}

func parseParam() *FlagParam {
	var fp FlagParam
	flag.StringVar(&fp.httpHost, "http-host", "127.0.0.1:54321", "对外http的地址，127.0.0.1:xxxx")
	flag.StringVar(&fp.mongoHost, "mongo-host", "192.168.1.199", "mongo地址")
	flag.StringVar(&fp.mongoPort, "mongo-port", "27017", "mongo端口")
	flag.StringVar(&fp.mongoUname, "mongo-name", "", "mongo用户名")
	flag.StringVar(&fp.mongoUpwd, "mongo-pwd", "", "mongo密码")
	flag.StringVar(&fp.mongoDb, "mongo-db", "lonlat_data", "mongo数据库名")
	flag.Parse()
	return &fp
}

var (
	db *mongodb.Mongodb
	fp = parseParam()
)

func init() {
	var cf mongodb.MongodbConf
	var err error
	cf.Db = fp.mongoDb
	cf.Host = fp.mongoHost
	cf.Port = fp.mongoPort
	cf.UName = fp.mongoUname
	cf.Upwd = fp.mongoUpwd
	db, err = mongodb.NewMongodb(cf)
	if err != nil {
		log.Fatal(err)
	}
}

func getGeop(w http.ResponseWriter, r *http.Request) {
	ad := strings.TrimSpace(r.URL.Query().Get("ad"))
	mlink, err := db.Get()
	lon := ""
	lat := ""

	if err != nil {
		log.Error(err)
		w.Write([]byte(""))
		return
	}

	defer mlink.Close()

	if ad == "" {
		w.WriteHeader(404)
		w.Write([]byte(""))
		return
	}

	qconf := mongodb.MongodbQueryConf{}
	qconf.Db = "lonlat_data"
	qconf.Table = "tbl_map"
	qconf.Select = mongodb.MM{"lon": 1, "lat": 1}
	qconf.Query = mongodb.MM{"ad": ad}
	info, err := mlink.One(qconf)
	if err != nil {
		log.Error(err)
		w.Write([]byte(""))
		return
	}

	if v, ok := info["lon"]; ok {
		lon = convert.ToString(v)
	}
	if v, ok := info["lat"]; ok {
		lat = convert.ToString(v)
	}
	w.Write([]byte(fmt.Sprintf("%s,%s", lon, lat)))
}

func main() {
	http.HandleFunc("/", getGeop)
	log.Error(http.ListenAndServe(fp.httpHost, nil))
}

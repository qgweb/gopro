// 导入地图数据到elasticsearch
package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/ngaut/log"
	"github.com/qgweb/new/lib/encrypt"
	//"github.com/qgweb/new/lib/timestamp"
	"gopkg.in/olivere/elastic.v3"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

var (
	pm *Param
	es *elastic.Client
)

type Param struct {
	GeoHost  string
	ESHost   string
	Type     string
	FileName string
}

func ParseParam() *Param {
	var p = &Param{}
	flag.StringVar(&p.ESHost, "host", "http://192.168.1.218:9200,http://192.168.1.218:9201", "es host")
	flag.StringVar(&p.Type, "type", "0", "0 电商搜索， 1 搜索引擎搜索")
	flag.StringVar(&p.FileName, "path", "", "处理文件绝对路径名称")
	flag.StringVar(&p.GeoHost, "ghost", "http://127.0.0.1:54321", "获取经纬度地址")
	flag.Parse()
	return p
}

func GetLonLat(ad string) string {
	r, err := http.Get(fmt.Sprintf("%s/?ad=%s", pm.GeoHost, ad))
	if err != nil {
		log.Error(err)
		return ""
	}

	if r != nil && r.Body != nil {
		defer r.Body.Close()
		v, _ := ioutil.ReadAll(r.Body)
		vs := strings.Split(string(v), ",")
		if len(vs) == 2 {
			return vs[1] + "," + vs[0]
		}
	}
	return ""
}

func init() {
	var err error
	pm = ParseParam()
	es, err = elastic.NewClient(elastic.SetURL(strings.Split(pm.ESHost, ",")...))
	if err != nil {
		log.Fatal(err)
	}
}

func getAdUa(v string, split string) (string, string) {
	info := strings.Split(v, split)
	if len(info) == 2 {
		return info[0], info[1]
	}
	return "", ""
}

func getKeyWord(v string, split string) []string {
	return strings.Split(v, split)
}

func BulkInsertData(reader *bufio.Reader, ktype string) {
	bk := es.Bulk()
	num := 0
	for {
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		}
		info := strings.Split(line, "\t#\t")
		log.Info(len(info))
		if len(info) != 2 {
			continue
		}
		ad, ua := getAdUa(info[0], "\t")
		keyword := getKeyWord(info[1], "\t")
		lonlat := GetLonLat(ad)
		if lonlat == "" {
			continue
		}
		num++
		id := encrypt.DefaultMd5.Encode("1456185600" + ad + ua)
		pinfo := map[string]interface{}{
			"ad":  ad,
			"ua":  ua,
			ktype: keyword,
			"geo": lonlat,
		}
		bk.Add(elastic.NewBulkUpdateRequest().Index("map_trace").Type("map").Doc(pinfo).Id(id).DocAsUpsert(true))
		bk.Add(elastic.NewBulkUpdateRequest().Index("map_trace_search").Type("map").Doc(pinfo).Id(id).DocAsUpsert(true))
		if num%10000 == 0 {
			log.Error(bk.Do())
		}
	}
	log.Info(bk.Do())
}

func main() {
	f, err := os.Open(pm.FileName)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	bi := bufio.NewReader(f)
	tt := ""
	if pm.Type == "0" {
		tt = "shopping_search"
	}
	if pm.Type == "1" {
		tt = "engine_search"
	}
	BulkInsertData(bi, tt)
}

package model

import (
	"flag"
	"fmt"
	"github.com/ngaut/log"
	"github.com/nsqio/go-nsq"
	"net/http"
	"io/ioutil"
	"strings"
)

type DataStorer interface {
	Receive(string, string, string)
}

type Storer interface {
	ParseData(interface{}) interface{}
	Save(interface{})
}

type DataStore struct {
	sr Storer
}

func (this *DataStore) HandleMessage(m *nsq.Message) error {
	if m.Body == nil {
		return nil
	}
	this.sr.Save(this.sr.ParseData(m.Body))
	return nil
}

func (this *DataStore) Receive(rkey string, host string, port string) {
	cus, err := nsq.NewConsumer(rkey, "goods", nsq.NewConfig())
	if err != nil {
		log.Fatal("连接nsq失败,错误信息为:", err)
	}

	cus.AddHandler(this)
	log.Info(cus.ConnectToNSQD(fmt.Sprintf("%s:%s", host, port)))

	for {
		select {
		case <-cus.StopChan:
			return
		}
	}
}

type Config struct {
	NsqHost       string
	NsqPort       string
	HbaseHost     string
	HbasePort     string
	MgoStoreHost  string
	MgoStorePort  string
	mgoStoreUname string
	mgoStoreUpwd  string
	MgoPutHost    string
	MgoPutPort    string
	ReceiveKey    string
	TablePrefixe  string //*前缀无_
	ESHost        string
	GType         string //数据源类型，淘宝，京东
	GeoHost       string
}

func ParseConfig() (cg Config) {
	flag.StringVar(&cg.NsqHost, "nsq-host", "127.0.0.1", "nsq 地址")
	flag.StringVar(&cg.NsqPort, "nsq-port", "4150", "nsq 端口")
	flag.StringVar(&cg.HbaseHost, "hbase-host", "192.168.1.218", "hbase 地址")
	flag.StringVar(&cg.HbasePort, "hbase-port", "2181", "hbase 端口")
	flag.StringVar(&cg.ReceiveKey, "rKey", "", "receive key")
	flag.StringVar(&cg.MgoStoreHost, "mdb-store-host", "192.168.1.199", "mongodb地址")
	flag.StringVar(&cg.MgoStorePort, "mdb-store-port", "27017", "mongodb端口")
	flag.StringVar(&cg.mgoStoreUname, "mdb-store-uname", "", "mongodb用户名")
	flag.StringVar(&cg.mgoStoreUpwd, "mdb-store-upwd", "", "mongodb用户密码")
	flag.StringVar(&cg.MgoPutHost, "mdb-put-host", "192.168.1.199", "mongodb地址")
	flag.StringVar(&cg.MgoPutPort, "mdb-put-port", "27017", "mongodb端口")
	flag.StringVar(&cg.TablePrefixe, "table_prefixe", "zhejiang_", "表前缀")
	flag.StringVar(&cg.ESHost, "es-host", "http://192.168.1.218:9200", "es地址多个，按逗号隔开")
	flag.StringVar(&cg.GType, "gtype", "taobao", "类型：taobao,jd")
	flag.StringVar(&cg.GeoHost, "geo-host", "http://127.0.0.1:54321", "经纬度转换地址")
	flag.Parse()
	return
}

func GetLonLat(ad string,host string) string {
	r, err := http.Get(fmt.Sprintf("%s/?ad=%s", host, ad))
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
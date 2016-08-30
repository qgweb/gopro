package lib

import (
	"bufio"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/qgweb/new/lib/config"
	"github.com/qgweb/new/lib/mongodb"
	"github.com/qgweb/new/lib/rediscache"
	"gopkg.in/olivere/elastic.v3"
)

var (
	configObj config.ConfigContainer
)

func init() {
	var err error
	if configObj, err = GetConfObj(GetConfigPath()); err != nil {
		log.Fatalln(err)
	}
}

// 获取程序执行目录
func GetBasePath() string {
	file, _ := exec.LookPath(os.Args[0])
	path, _ := filepath.Abs(file)
	return filepath.Dir(path)
}

// 获取配置文件对象
func GetConfigPath() string {
	return GetBasePath() + "/conf.ini"
}

// 获取配置文件对象
func GetConfObj(iniPath string) (config.ConfigContainer, error) {
	return config.NewConfig("ini", iniPath)
}

// 获取配置文件节点值
func GetConfVal(key string) string {
	return configObj.String(key)
}

// 统计数据
func StatisticsData(db, key, val, opt string) error {
	client := http.Client{
		Transport: &http.Transport{
			Dial: func(netw, addr string) (net.Conn, error) {
				deadline := time.Now().Add(10 * time.Second)
				c, err := net.DialTimeout(netw, addr, time.Second * 10)
				if err != nil {
					return nil, err
				}
				c.SetDeadline(deadline)
				return c, nil
			},
		},
	}
	surl := GetConfVal("default::stats_url")
	v := url.Values{}
	v.Set("db", db)
	v.Set("key", key)
	v.Set("value", val)
	v.Set("opt", opt)

	res, err := client.Post(surl + "api/create", "application/x-www-form-urlencoded",
		ioutil.NopCloser(strings.NewReader(v.Encode())))
	if err != nil {
		return err
	}

	if res != nil && res.Body != nil {
		res.Body.Close()
	}
	return nil
}

// 获取es对象
func GetESObj() (*elastic.Client, error) {
	surl := GetConfVal("default::es_url")
	es, err := elastic.NewClient(elastic.SetURL(strings.Split(surl, ",")...))
	return es, err
}

// 获取filedb数据
func GetFdbData(fname string, fun func(string)) error {
	furl := GetConfVal("default::fdb_url")
	res, err := http.Get(furl + "?name=" + fname)
	if err != nil {
		return err
	}
	if res != nil && res.Body != nil {
		bi := bufio.NewReader(res.Body)
		for {
			line, err := bi.ReadString('\n')
			if err == io.EOF {
				break
			}
			if err != nil {
				continue
			}

			fun(strings.TrimSpace(line))
		}
	}
	return nil
}

// 获取redis对象
func GetRedisObj() (*rediscache.MemCache, error) {
	rurls := strings.Split(GetConfVal("default::put_redis_url"), ":")
	conf := rediscache.MemConfig{}
	conf.Host = rurls[0]
	conf.Port = rurls[1]
	return rediscache.New(conf)
}

// 获取redis对象
func GetPutRedisObj(key string) (conn *rediscache.MemCache, err error) {
	rurls := strings.Split(GetConfVal("default::" + key), ":")
	conf := rediscache.MemConfig{}
	conf.Host = rurls[0]
	conf.Port = rurls[1]
	conf.Db = rurls[2]
	conn, err = rediscache.New(conf)
	conn.SelectDb(conf.Db)
	return
}

// 获取redis对象
func GetZJDxRedisObj() (*rediscache.MemCache, error) {
	rurls := strings.Split(GetConfVal("zhejiang::dx_redis_url"), ":")
	conf := rediscache.MemConfig{}
	conf.Host = rurls[0]
	conf.Port = rurls[1]
	return rediscache.New(conf)
}

// 获取mongo对象
func GetMongoObj() (*mongodb.Mongodb, error) {
	murls := strings.Split(GetConfVal("zhejiang::mongo_url"), ":")
	conf := mongodb.MongodbConf{}
	conf.Host = murls[0]
	conf.Port = murls[1]
	conf.Db = "data_source"
	return mongodb.NewMongodb(conf)
}

// 数据添加前缀
func AddPrefix(val string, pre string) string {
	var list = strings.Split(val, "\t")
	if len(list) != 3 {
		return ""
	}
	list[2] = pre + strings.Join(strings.Split(list[2], ","), "," + pre)
	return strings.Join(list, "\t")
}

// 数据添加前缀
func AddPrefix2(val string, pre string) string {
	var nlist = make([]string, 3)
	var list = strings.Split(val, "\t")
	if len(list) != 2 {
		return ""
	}
	nlist[0] = list[0]
	nlist[1] = "ua"
	nlist[2] = list[1]
	nlist[2] = pre + strings.Join(strings.Split(nlist[2], ","), "," + pre)
	return strings.Join(nlist, "\t")
}

// 是否是mongoid
func IsMongo(val string) bool {
	return mongodb.IsObjectId(val)
}

// rtb
package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/garyburd/redigo/redis"
	"github.com/qgweb/gopro/lib/encrypt"
	"github.com/qgweb/new/lib/config"
	"io"
	"log"
	"net/http"
	_ "net/http/pprof"
	"net/url"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"
)

var (
	conf      = flag.String("conf", "", "配置文件")
	IniFile   config.ConfigContainer
	err       error
	pool      *redis.Pool
	aupool    *redis.Pool // 记录投放过的adua
	rtotal    uint64
	stotal    uint64
	adtotal   uint64
	aduatotal uint64
	fileChan  chan string
)

const (
	MODEL_REDIS = 1
	MODEL_URL   = 2
)

func init() {
	flag.Parse()

	if *conf == "" {
		log.Fatalln("配置文件不存在")
	}

	IniFile, err = config.NewConfig("ini", *conf)
	if err != nil {
		log.Fatalln("读取配置文件出错,错误信息为:", err)
	}

	pool = GetRedisPool(IniFile.String("redis::host"),
		IniFile.String("redis::port"))
	aupool = GetRedisPool(IniFile.String("auredis::host"),
		IniFile.String("auredis::port"))
	fileChan = make(chan string, 1000)
}

func main() {
	var (
		host = IniFile.String("http::host")
		port = IniFile.String("http::port")
	)

	if host == "" || port == "" {
		log.Fatalln("电信接口配置出错")
	}
	go recordBiddingRuquest()
	runtime.GOMAXPROCS(runtime.NumCPU())
	http.HandleFunc("/uri", requestPrice)
	http.HandleFunc("/tj", func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "recv-total:"+strconv.FormatUint(rtotal, 10)+"\n")
		io.WriteString(w, "deal-total:"+strconv.FormatUint(stotal, 10)+"\n")
		io.WriteString(w, "adua-total:"+strconv.FormatUint(aduatotal, 10)+"\n")
		io.WriteString(w, "ad-total:"+strconv.FormatUint(adtotal, 10)+"\n")
	})
	log.Println(http.ListenAndServe(fmt.Sprintf("%s:%s", host, port), nil))
}

//请求出价
func requestPrice(w http.ResponseWriter, r *http.Request) {
	var isopen = IniFile.String("default::open")
	rtotal++
	//http://$ip:$port/uri/?ad=$ad&ua=$ua&url=$url&mid=$mid&showType=01|02|03|04|05&price=$price
	query := r.URL.Query()
	ad := query.Get("ad")
	ua := query.Get("ua")
	//	url := query.Get("url")
	price := query.Get("price")

	mid := query.Get("mid")
	//	showType := query.Get("showType")

	fileChan <- fmt.Sprintf("%s\t%s\t%s", ad, ua, price)

	if mid == "" {
		w.WriteHeader(404)
		return
	}

	param := make(map[string]string)
	for k, v := range query {
		param[k] = v[0]
	}

	if isopen == "1" {
		go reponsePrice(param)
	}
}

//请求出价
func reponsePrice(param map[string]string) {
	//	mux.Lock()
	//	defer mux.Unlock()

	//	//http://ip:port/uri?mid=$mid&prod=$prod&showType=$showType
	var (
		host  = IniFile.String("dxhttp::host")
		port  = IniFile.String("dxhttp::port")
		adurl = IniFile.String("dxhttp::url")
		mode  = IniFile.String("default::mode")
	)

	if host == "" || port == "" {
		log.Fatalln("读取电信配置文件出错")
	}

	if _, ok := param["ad"]; ok {
		adurl = adurl + "?sh_cox=" + param["ad"]
	}

	url := fmt.Sprintf("http://%s:%s/receive?mid=%s&prod=%s&showType=%s&token=%s&price=%s",
		host, port, param["mid"], encrypt.GetEnDecoder(encrypt.TYPE_BASE64).Encode(adurl),
		"03", "reBkYQmESMs=", "10")

	if checkExistAd(param["ad"]) {
		return
	}

	isput := false
	switch mode {
	case MODEL_REDIS:
		isput = matchRedis(param) //redis匹配
	case MODEL_URL:
		isput = matchUrl(param) //url匹配
	default: //所有
		isput = matchRedis(param) || matchUrl(param)
	}

	if !isput {
		return
	}

	resp, err := http.Get(url)
	stotal++
	if resp != nil {
		resp.Body.Close()
	}
	if err != nil {
		log.Println("发送数据出错,错误信息为:", err)
		return
	}
	if resp != nil && resp.StatusCode != 200 {
		log.Println("返回数据出错,错误code为:", resp.StatusCode)
		return
	}

	recordPutAd(param["ad"])
}

// 匹配redis
func matchRedis(param map[string]string) bool {
	conn := pool.Get()
	defer conn.Close()

	conn.Do("SELECT", "1")

	key := encrypt.DefaultMd5.Encode(param["ad"] + "_" + param["ua"])
	r, err := redis.Bool(conn.Do("EXISTS", key))
	if err != nil || !r {
		return false
		r, err = redis.Bool(conn.Do("EXISTS", param["ad"]))
		if err != nil || !r {
			return false
		} else {
			adtotal++
		}
	} else {
		aduatotal++
	}
	return true
}

// 解析url地址
func parseUrl(ourl string) string {
	if !(strings.Contains(ourl, "http://") || strings.Contains(ourl, "https://")) {
		ourl = "//" + ourl
	}
	a, err := url.Parse(ourl)
	if err != nil {
		return ""
	}
	return a.Host + a.Path
}

// 匹配链接
func matchUrl(param map[string]string) bool {
	var db = IniFile.String("auredis::urldb")
	var key = parseUrl(encrypt.DefaultBase64.Decode(param["url"]))
	conn := aupool.Get()
	defer conn.Close()

	conn.Do("SELECT", db)
	v, _ := redis.Bool(conn.Do("EXISTS", key))
	return v
}

// 验证是否存在
func checkExistAd(ad string) bool {
	var db = IniFile.String("auredis::db")
	var key = "SH_HPUT_" + time.Now().Format("20060102")
	conn := aupool.Get()
	defer conn.Close()

	conn.Do("SELECT", db)

	if r, _ := redis.Bool(conn.Do("HEXISTS", key, ad)); r {
		return true
	}
	return false
}

// 记录投放过的ad
func recordPutAd(ad string) {
	var db = IniFile.String("auredis::db")
	var key = "SH_HPUT_" + time.Now().Format("20060102")
	conn := aupool.Get()
	defer conn.Close()

	conn.Do("SELECT", db)
	conn.Do("HSET", key, ad, 1)
}

//redis链接此
func GetRedisPool(host, port string) *redis.Pool {
	return &redis.Pool{
		MaxIdle:   100,
		MaxActive: 100, // max number of connections
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", host+":"+port)
			if err != nil {
				panic(err.Error())
			}
			return c, err
		},
	}
}

func recordBiddingRuquest() {
	f, err := os.OpenFile("./adua.txt", os.O_APPEND|os.O_CREATE|os.O_RDWR, os.ModePerm)
	if err != nil {
		log.Fatalln(err)
	}
	defer f.Close()
	bw := bufio.NewWriter(f)

	for {
		select {
		case msg := <-fileChan:
			bw.WriteString(msg + "\n")
			break
		}
	}
	bw.Flush()
}

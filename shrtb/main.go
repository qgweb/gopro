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
	"math/rand"
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
	MODEL_REDIS = "1"
	MODEL_URL   = "2"
)

// 物料结构
type materials struct {
	isput  bool   //是否有物料
	puturl string //投放地址
	uniqId string //唯一id，用于频次控制
}

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
	url := query.Get("url")
	price := query.Get("price")
	mid := query.Get("mid")
	//	showType := query.Get("showType")

	fileChan <- fmt.Sprintf("%s\t%s\t%s\t%s", ad, ua, price,
		encrypt.DefaultBase64.Decode(url))

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
	//	//http://ip:port/uri?mid=$mid&prod=$prod&showType=$showType
	var (
		mode = IniFile.String("default::mode")
	)

	param["DX_HOST"] = IniFile.String("dxhttp::host")
	param["DX_PORT"] = IniFile.String("dxhttp::port")
	param["XU_URL"] = IniFile.String("adurl::url")
	if param["DX_HOST"] == "" || param["DX_PORT"] == "" {
		log.Fatalln("读取电信配置文件出错")
	}

	pm := materials{}
	switch mode {
	case MODEL_REDIS:
		pm = matchRedis(param) //redis匹配
	case MODEL_URL:
		pm = matchUrl(param) //url匹配
	default: //所有
		var pms = make([]materials, 0, 2)
		if ms := matchRedis(param); ms.isput {
			pms = append(pms, ms)
		}
		if ms := matchUrl(param); ms.isput {
			pms = append(pms, ms)
		}
		pm = pms[randNum(len(pms))]
	}
	log.Println(pm)
	if !pm.isput {
		return
	}

	if checkExistAd(param["ad"], pm.uniqId) {
		return
	}

	resp, err := http.Get(pm.puturl)
	stotal++
	if resp != nil {
		resp.Body.Close()
	}
	if err != nil {
		log.Println("发送数据出错,错误信息为:", err)
		//return
	}
	if resp != nil && resp.StatusCode != 200 {
		log.Println("返回数据出错,错误code为:", resp.StatusCode)
		//return
	}

	recordPutAd(param["ad"], pm.uniqId)
}

// 匹配redis
func matchRedis(param map[string]string) (ms materials) {
	conn := pool.Get()
	defer conn.Close()

	conn.Do("SELECT", "1")
	key := param["ad"]
	aids, err := redis.Strings(conn.Do("SMEMBERS", key))
	if err != nil || len(aids) == 0 {
		return ms
	}
	aid := aids[randNum(len(aids))]
	adurl := fmt.Sprintf("%s?sh_cox=%s&aid=%s", param["XU_URL"], param["ad"], aid)
	purl := fmt.Sprintf("http://%s:%s/receive?mid=%s&prod=%s&showType=%s&token=%s&price=%s",
		param["DX_HOST"], param["DX_PORT"], param["mid"], encrypt.GetEnDecoder(encrypt.TYPE_BASE64).Encode(adurl),
		"03", "reBkYQmESMs=", "10")
	return materials{isput: true, puturl: purl, uniqId: aid}
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
	return a.Host
}

// 随机区一个
func randNum(size int) int {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return r.Intn(size)
}

// 匹配链接
func matchUrl(param map[string]string) (ms materials) {
	var db = IniFile.String("auredis::urldb")
	var key = parseUrl(encrypt.DefaultBase64.Decode(param["url"]))
	conn := aupool.Get()
	defer conn.Close()

	conn.Do("SELECT", db)
	v, _ := redis.Strings(conn.Do("SMEMBERS", key))

	if len(v) == 0 {
		return ms
	}

	//出价，尺寸，素材地址
	urls := strings.Split(v[randNum(len(v))], "\t")
	if len(urls) < 3 {
		return ms
	}
	purl := fmt.Sprintf("http://%s:%s/receive?mid=%s&prod=%s&showType=%s&token=%s&price=%s",
		param["DX_HOST"], param["DX_PORT"], param["mid"], encrypt.GetEnDecoder(encrypt.TYPE_BASE64).Encode(urls[2]),
		urls[1], "reBkYQmESMs=", urls[0])
	return materials{isput: true, puturl: purl, uniqId: urls[2]}
}

// 验证是否存在
func checkExistAd(ad string, url string) bool {
	var db = IniFile.String("auredis::db")
	var key = "SH_HPUT_" + time.Now().Format("20060102")
	var dkey = encrypt.DefaultMd5.Encode(ad + "_" + url)
	conn := aupool.Get()
	defer conn.Close()

	conn.Do("SELECT", db)

	if r, _ := redis.Bool(conn.Do("HEXISTS", key, dkey)); r {
		return true
	}
	return false
}

// 记录投放过的ad
func recordPutAd(ad string, url string) {
	var db = IniFile.String("auredis::db")
	var key = "SH_HPUT_" + time.Now().Format("20060102")
	var dkey = encrypt.DefaultMd5.Encode(ad + "_" + url)
	conn := aupool.Get()
	defer conn.Close()
	conn.Do("SELECT", db)
	conn.Do("HSET", key, dkey, 1)
}

//redis链接此
func GetRedisPool(host, port string) *redis.Pool {
	return &redis.Pool{
		MaxIdle:   100,
		MaxActive: 100, // max number of connections
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", host+":"+port)
			if err != nil {
				return nil, err
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
			bw.WriteString(time.Now().Format("2006-01-02") + "\t" + msg + "\n")
			break
		}
	}
	bw.Flush()
}

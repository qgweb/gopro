// rtb
package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/goweb/gopro/lib/convert"
	"github.com/goweb/gopro/lib/encrypt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"runtime"
	"strconv"
	"sync"
	"time"

	_ "net/http/pprof"

	"github.com/garyburd/redigo/redis"
	"gopkg.in/ini.v1"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var (
	conf       = flag.String("conf", "", "配置文件")
	IniFile    *ini.File
	err        error
	pool       *redis.Pool
	httpclient http.Client
	mux        sync.Mutex
	rtotal     uint64
	stotal     uint64
	mdbsession *mgo.Session
)

func init() {
	flag.Parse()

	if *conf == "" {
		log.Fatalln("配置文件不存在")
	}

	confData, err := ioutil.ReadFile(*conf)
	if err != nil {
		log.Fatalln("读取配置文件出错,错误信息为:", err)
	}

	IniFile, err = ini.Load(confData)
	if err != nil {
		log.Fatalln("生成配置文件结构出错,错误信息为:", err)
	}

	pool = GetRedisPool()

	//	httpclient = http.Client{
	//		Transport: &http.Transport{
	//			Dial: func(netw, addr string) (net.Conn, error) {
	//				deadline := time.Now().Add(25 * time.Second)
	//				c, err := net.DialTimeout(netw, addr, time.Second*20)
	//				if err != nil {
	//					return nil, err
	//				}
	//				c.SetDeadline(deadline)
	//				return c, nil
	//			},
	//			DisableKeepAlives: true,
	//		},
	//	}
}

func main() {
	var (
		host = IniFile.Section("http").Key("host").String()
		port = IniFile.Section("http").Key("port").String()
	)

	if host == "" || port == "" {
		log.Fatalln("电信接口配置出错")
	}

	runtime.GOMAXPROCS(runtime.NumCPU())
	http.HandleFunc("/uri", requestPrice)
	http.HandleFunc("/tj", func(w http.ResponseWriter, r *http.Request) {

		io.WriteString(w, "recv-total:"+strconv.FormatUint(rtotal, 10))
		io.WriteString(w, "deal-total:"+strconv.FormatUint(stotal, 10))
	})
	log.Println(http.ListenAndServe(fmt.Sprintf("%s:%s", host, port), nil))
}

func getRand(ary []int) int {
	prosum := 0
	for _, v := range ary {
		prosum += v
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	for k, v := range ary {
		randNum := r.Intn(prosum)
		if randNum <= v {
			return k
		} else {
			prosum -= v
		}
	}

	return -1
}

//请求出价
func requestPrice(w http.ResponseWriter, r *http.Request) {
	rtotal++
	//http://$ip:$port/uri/?ad=$ad&ua=$ua&url=$url&mid=$mid&showType=01|02|03|04
	query := r.URL.Query()
	//	ad := query.Get("ad")
	//	ua := query.Get("ua")
	//	url := query.Get("url")
	mid := query.Get("mid")
	//	showType := query.Get("showType")

	if mid == "" {
		w.WriteHeader(404)
		return
	}

	param := make(map[string]string)
	for k, v := range query {
		param[k] = v[0]
	}

	go recordAdUa(param)

	//30%概率返回
	rd := getRand([]int{30, 70})
	if rd == 0 {
		go reponsePrice(param)
	}
}

//请求出价
func reponsePrice(param map[string]string) {
	//	mux.Lock()
	//	defer mux.Unlock()

	//	//http://ip:port/uri?mid=$mid&prod=$prod&showType=$showType
	//	conn := pool.Get()
	//	defer conn.Close()

	var (
		host  = IniFile.Section("dxhttp").Key("host").String()
		port  = IniFile.Section("dxhttp").Key("port").String()
		adurl = IniFile.Section("adurl").Key("url").String()
	)

	if host == "" || port == "" {
		log.Fatalln("读取电信配置文件出错")
	}

	if _, ok := param["ad"]; ok {
		adurl = adurl + "?cox=" + param["ad"]
	}

	url := fmt.Sprintf("http://%s:%s/receive?mid=%s&prod=%s&showType=%s&token=%s",
		host, port, param["mid"], encrypt.GetEnDecoder(encrypt.TYPE_BASE64).Encode(adurl),
		"03", "reBkYQmESMs=")

	//	res, err := redis.Bytes(conn.Do("GET", "name"))
	//	if err != nil {
	//		log.Println(err)
	//		return
	//	}

	//	if string(res) == "" {
	//		fmt.Println(res)
	//	}

	//resp, err := httpclient.Get(url)
	resp, err := http.Get(url)

	if err != nil {
		log.Println("发送数据出错,错误信息为:", err)
	}
	if resp != nil && resp.StatusCode != 200 {
		log.Println("返回数据出错,错误code为:", resp.StatusCode)
	}
	stotal++
}

//获取关键字信息
func getKeyWord(ad string) (string, error) {
	defer func() {
		if a := recover(); a != nil {
			log.Println(a)
		}
	}()

	tookenUrl := "http://61.129.39.71/telecom-dmp/getToken?apiKey=6c2d70c079160a60e9a91d7c514e3903&sign=dc1367fb302be81365de1dd20a1e4ffb3c1d81fd"
	resp, err := http.Get(tookenUrl)
	if err != nil {
		log.Println("获取token失败,错误信息:", err)
		return "", err
	}
	defer resp.Body.Close()

	tooken, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("读取tooken内容失败,错误信息:", err)
		return "", err
	}

	var tookenJson map[string]interface{}

	err = json.Unmarshal(tooken, &tookenJson)
	if err != nil {
		log.Println("解析json失败,错误信息:", err)
		return "", err
	}

	tookenStr := ""

	if convert.ToString(tookenJson["code"]) == "200200" &&
		convert.ToString(tookenJson["message"]) == "OK" {
		tookenStr = convert.ToString(tookenJson["result"])
	} else {
		log.Println("tooken获取失败")
		return "", errors.New("tooken获取失败")
	}

	keyUrl := fmt.Sprintf("http://61.129.39.71/telecom-dmp/kv/getValueByKey?token=%s&table=qg_ds_kv&key=%s", tookenStr, ad)

	resp, err = http.Get(keyUrl)
	if err != nil {
		log.Println("获取关键字信息失败,错误信息:", err)
		return "", err
	}
	defer resp.Body.Close()

	key, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("读取关键字信息内容失败,错误信息:", err)
		return "", err
	}

	var keyJson map[string]interface{}

	err = json.Unmarshal(key, &keyJson)
	if err != nil {
		log.Println("解析json失败,错误信息:", err)
		return "", err
	}

	if convert.ToString(keyJson["code"]) == "200200" &&
		convert.ToString(keyJson["message"]) == "OK" &&
		keyJson["result"] != nil {
		return convert.ToString(keyJson["result"].(map[string]interface{})["value"]), nil
	} else {
		log.Println("关键字获取失败")
		return "", errors.New("关键字获取失败")
	}
}

//记录信息
func recordAdUa(param map[string]string) {
	sess := GetSession()
	defer sess.Close()

	db := IniFile.Section("mongo").Key("db").String()

	res, err := getKeyWord(param["ad"])
	if err != nil {
		res = ""
	}

	res = encrypt.GetEnDecoder(encrypt.TYPE_BASE64).Decode(res)

	sess.DB(db).C("srtb").Insert(bson.M{"ad": param["ad"], "ua": encrypt.GetEnDecoder(encrypt.TYPE_BASE64).Decode(param["ua"]),
		"keys": res, "date": time.Now().Format("2006-01-02")})
}

//redis链接此
func GetRedisPool() *redis.Pool {
	var (
		host = IniFile.Section("redis").Key("host").String()
		port = IniFile.Section("redis").Key("port").String()
	)

	return &redis.Pool{
		MaxIdle:   10240,
		MaxActive: 10240, // max number of connections
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", host+":"+port)
			if err != nil {
				panic(err.Error())
			}
			return c, err
		},
	}
}

//获取mongo-session
func GetSession() *mgo.Session {
	mux.Lock()
	defer mux.Unlock()

	if mdbsession == nil {
		var (
			err    error
			mouser = IniFile.Section("mongo").Key("user").String()
			mopwd  = IniFile.Section("mongo").Key("pwd").String()
			mohost = IniFile.Section("mongo").Key("host").String()
			moport = IniFile.Section("mongo").Key("port").String()
			modb   = IniFile.Section("mongo").Key("db").String()
		)
		fmt.Println(fmt.Sprintf("%s:%s@%s:%s/%s", mouser, mopwd, mohost, moport, modb))
		mdbsession, err = mgo.Dial(fmt.Sprintf("%s:%s@%s:%s/%s", mouser, mopwd, mohost, moport, modb))
		if err != nil {
			panic(err)
		}
	}
	//高并发下会关闭连接,ping下会恢复
	mdbsession.Ping()
	return mdbsession.Copy()
}

//处理数据-抓取数据-存标签库
package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/Unknwon/goconfig"
	"github.com/garyburd/redigo/redis"

	"goclass/encrypt"

	"github.com/astaxie/beego/httplib"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var (
	conf     = flag.String("conf", "conf.ini", "配置文件")
	rhost    = ""
	rport    = ""
	mohost   = ""
	moport   = ""
	mouser   = ""
	mopwd    = ""
	modb     = ""
	qshost   = ""
	qsport   = ""
	qsname   = ""
	qsauth   = ""
	cookie   = ""
	rediskey = ""
	prefix   = ""

	mdbsession *mgo.Session = nil
	mux        sync.Mutex
	err        error
	iniFile    *goconfig.ConfigFile
	rpool      *redis.Pool
	sqsClient  *httplib.BeegoHttpRequest
)

func init() {
	//解析参数
	flag.Parse()

	if *conf == "" {
		log.Fatalln("配置文件不能为空")
	}

	//读取参数
	iniFile, err := goconfig.LoadConfigFile(*conf)
	if err != nil {
		log.Fatalln("读取配置文件失败,", err)
	}

	mohost, err = iniFile.GetValue("mongo", "host")
	if err != nil {
		log.Fatalln("读取mongo-host失败:", err)
	}
	moport, err = iniFile.GetValue("mongo", "port")
	if err != nil {
		log.Fatalln("读取mongo-port失败:", err)
	}
	modb, err = iniFile.GetValue("mongo", "db")
	if err != nil {
		log.Fatalln("读取mongo-db失败:", err)
	}
	mouser, err = iniFile.GetValue("mongo", "user")
	if err != nil {
		log.Fatalln("读取mongo-user失败:", err)
	}
	mopwd, err = iniFile.GetValue("mongo", "pwd")
	if err != nil {
		log.Fatalln("读取mongo-pwd失败:", err)
	}

	//httpsqs
	qsport, err = iniFile.GetValue("httpsqs", "port")
	if err != nil {
		log.Fatalln("读取httpsqs-port失败:", err)
	}
	qshost, err = iniFile.GetValue("httpsqs", "host")
	if err != nil {
		log.Fatalln("读取httpsqs-host失败:", err)
	}
	qsauth, err = iniFile.GetValue("httpsqs", "auth")
	if err != nil {
		log.Fatalln("读取httpsqs-auth失败:", err)
	}

	//cookie
	cookie, err = iniFile.GetValue("default", "cookie")
	if err != nil {
		log.Fatalln("读取cookie失败:", err)
	}

	//redis
	rport, err = iniFile.GetValue("redis", "port")
	if err != nil {
		log.Fatalln("读取redis-port失败:", err)
	}
	rhost, err = iniFile.GetValue("redis", "host")
	if err != nil {
		log.Fatalln("读取redis-host失败:", err)
	}

	rediskey, err = iniFile.GetValue("queuekey", "key")
	if err != nil {
		log.Fatalln("读取队列key失败", err)
	}

	prefix, err = iniFile.GetValue("queuekey", "prefix")
	if err != nil {
		log.Fatalln("读取队列前缀失败", err)
	}

	//rpool = RedisPool()
}

//获取mongo数据库链接
func GetSession() *mgo.Session {
	mux.Lock()
	defer mux.Unlock()

	if mdbsession == nil {
		var err error
		mdbsession, err = mgo.Dial(fmt.Sprintf("%s:%s@%s:%s/%s", mouser, mopwd, mohost, moport, modb))
		if err != nil {
			panic(err)
		}
	}
	//高并发下会关闭连接,ping下会恢复
	mdbsession.Ping()
	return mdbsession.Copy()
}

//获取reids连接池
func RedisPool() *redis.Pool {
	return &redis.Pool{
		MaxIdle:   5,
		MaxActive: 10, // max number of connections
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", rhost+":"+rport)
			if err != nil {
				log.Fatalln(err.Error())
			}
			return c, err
		},
	}
}

//redis queue
func redisQueue(px string) url.Values {
	conn := rpool.Get()
	defer conn.Close()

	//res, err := redis.Strings(conn.Do("BLPOP", rediskey+px,"0"))
	res, err := redis.Bytes(conn.Do("LPOP", rediskey+px))
	if err != nil {
		//log.Println("读取redis数据失败:", err)
		time.Sleep(time.Minute * 10)
		return nil
	}

	if res == nil {
		return nil
	}

	data, err := url.ParseQuery(string(res))
	if err != nil {
		log.Println("解析数据失败")
		return nil
	}

	return data
}

// 队列操作
func httpsqsQueue(px string) url.Values {
	hurl := fmt.Sprintf("http://%s:%s/?name=%s&opt=%s&auth=%s", qshost, qsport,
		rediskey+px, "get", qsauth)

	r := httplib.Get(hurl)
	transport := http.Transport{
		DisableKeepAlives: true,
	}
	r.SetTransport(&transport)
	res, err := r.String()

	if err != nil {
		log.Println("读取http队列出错,错误信息为:", err)
		return nil
	}

	if string(res) == "HTTPSQS_GET_END" || string(res) == "HTTPSQS_ERROR" {
		return nil
	}

	res = encrypt.GetEnDecoder(encrypt.TYPE_BASE64).Decode(res)

	data, err := url.ParseQuery(res)
	if err != nil {
		log.Println("解析数据失败")
		return nil
	}

	return data
}

//程序入口
func Bootstrap(px string) {
	for {
		//data := redisQueue(px)
		data := httpsqsQueue(px)

		if data == nil {
			time.Sleep(time.Minute)
			continue
		}
		Dispath(data)
		time.Sleep(time.Millisecond * 500)
	}
}

//分配数据
func Dispath(data url.Values) {
	defer func() {
		if msg := recover(); msg != nil {
			log.Println(msg)
		}
	}()

	ad := data.Get("ad")
	cookieId := data.Get("cookie")
	ua := data.Get("ua")
	gids := strings.Split(strings.TrimSpace(data.Get("uids")), ",")
	cids := make([]string, 0, len(gids))
	shopids := make([]string, 0, len(gids))

	//多线程抓取
	goodsList := make(chan map[string]string, 20)
	overTag := make(chan bool)
	readCount := 0

	go func() {
		for _, gid := range gids {
			go func(gid string) {
				//判断是否存在
				ginfo, err := checkGoodsExist(gid)

				if err != mgo.ErrNotFound && err != nil {
					log.Println(err)
					goodsList <- nil
					return
				}
				if err == mgo.ErrNotFound {
					//抓取商品
					ginfo = AddGoodsInfo(gid)

					if ginfo == nil {
						goodsList <- nil
						return
					}
				}

				goodsList <- ginfo

				//添加店铺表
				if !checkShopExist(ginfo["shop_id"]) {
					go AddShopInfo(ginfo)
				}
			}(gid)
		}
	}()

	go func() {
		for {
			info := <-goodsList

			if info != nil && len(info) != 0 {
				cids = append(cids, info["cid"])
				shopids = append(shopids, info["shop_id"])
			}

			readCount++

			if readCount == len(gids) {
				overTag <- true
				break
			}
		}
	}()

	<-overTag

	//添加对应关系
	if len(cids) != 0 {
		ua = encrypt.GetEnDecoder(encrypt.TYPE_BASE64).Decode(ua)
		//添加商品对应各种关系表
		go AddUidCids(map[string]string{
			"ad": ad, "cookie": cookieId,
			"ua": ua, "cids": strings.Join(cids, ","),
			"shops": strings.Join(shopids, ","),
			"clock": data.Get("clock"), "date": data.Get("date")})
	}
}

//判断是否存在该商品
func checkGoodsExist(gid string) (res map[string]string, err error) {
	sess := GetSession()
	defer func() {
		sess.Close()
	}()

	info := make(map[string]string)

	err = sess.DB(modb).C("goods").Find(bson.M{"gid": gid}).
		Select(bson.M{"tagid": 1, "_id": 0, "shop_id": 1}).One(&info)

	if err == mgo.ErrNotFound {
		return nil, err
	}

	//更新count字段
	sess.DB(modb).C("goods").Update(bson.M{"gid": gid}, bson.M{"$inc": bson.M{"count": 1}})

	if _, ok := info["shop_id"]; !ok {
		return nil, mgo.ErrNotFound
	}

	info1 := make(map[string]string)

	//获取店铺信息
	sess.DB(modb).C("taoshop").Find(bson.M{"shop_id": info["shop_id"]}).
		Select(bson.M{"_id": 0}).One(&info1)

	if _, ok := info1["shop_id"]; !ok {
		return nil, mgo.ErrNotFound
	}

	info1["tagid"] = info["tagid"]

	return info1, nil
}

//验证店铺是否存在
func checkShopExist(shopid string) bool {
	sess := GetSession()
	defer sess.Close()
	n, err := sess.DB(modb).C("taoshop").Find(bson.M{"shop_id": shopid}).Count()
	if err == nil && n > 0 {
		return true
	}
	return false
}

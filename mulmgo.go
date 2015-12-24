package main

import (
	"fmt"
	"github.com/garyburd/redigo/redis"
	"github.com/ngaut/log"
	"github.com/qgweb/gopro/lib/rediscache"
	"github.com/qgweb/gopro/lib/mongodb"
	"goclass/convert"
	"gopkg.in/mgo.v2/bson"
	"strings"
	"sync"
	"time"
)

func main() {
	config := rediscache.MemConfig{}
	config.Host = "127.0.0.1"
	config.Port = "6379"
	ldb, err := rediscache.New(config)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer ldb.Close()

	ldb.Set("aaa","123")
	fmt.Println(ldb.Get("aaa"))
	ldb.Del("aaa")
	fmt.Println(ldb.Get("aaa"))
	ldb.HSet("bbb","11","111")
	ldb.HSet("bbb","21","2222")
	fmt.Println(ldb.HGetAllKeys("bbb"))
	fmt.Println(ldb.HGetAllValue("bbb"))
	//ldb.Del("bbb")
	ldb.HDel("bbb","11")
	fmt.Println(ldb.HGetAllValue("bbb"))
	fmt.Println(ldb.Keys("*"))
	ldb.Expire("bbb",30)
	ldb.Flush()
	return


	c, err := mongodb.Dial("192.168.1.199:27017/data_source", 10)
	if err != nil {
		log.Error(err)
		return
	}
	defer c.Close()
	c.Debug()

	mux := sync.Mutex{}
	p := mongodb.MulQueryParam{}
	p.DbName = "data_source"
	p.ColName = "useraction"
	//	p.Query = bson.M{"timestamp": "1449417600"}
	p.Query = bson.M{}
	p.Size = 500000
	p.Fun = func(info map[string]interface{}) {
		mux.Lock()
		ad := info["AD"].(string)
		ua := info["UA"].(string)
		ldb.Set(ad+ua, "1")

		mux.Unlock()
	}
	bt := time.Now()
	fmt.Println(c.Query(p))
	ldb.Flush()
	fmt.Println(time.Now().Sub(bt).Seconds())

	return
	cc, err := redis.Dial("tcp", "192.168.1.199:6380")
	if err != nil {
		fmt.Println(err)
		return
	}

	cc.Do("SELECT", "10")
	if l := ldb.Keys("xxxx_ad_"); true {
		b1 := time.Now()
		for ii := 0; ii < 10; ii++ {
			for _, v := range l {
				k := strings.TrimLeft(v, "xxxx_ad_")
				cc.Send("SET", k+convert.ToString(ii), 1)
			}
		}

		fmt.Println(len(l))

		cc.Flush()
		fmt.Println(cc.Receive())
		fmt.Println(time.Now().Sub(b1).Seconds())
	}

}

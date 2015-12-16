package main

import (
	"fmt"
	"github.com/garyburd/redigo/redis"
	"github.com/ngaut/log"
	"github.com/qgweb/gopro/lib/cache"
	"github.com/qgweb/gopro/lib/mongodb"
	"goclass/convert"
	"gopkg.in/mgo.v2/bson"
	"strings"
	"time"
)

func main() {
	ldb, err := cache.NewLevelDbCache("/tmp/lldb")
	if err != nil {
		fmt.Print(err)
		return
	}
	defer ldb.Close()

	c, err := mongodb.Dial("192.168.1.199:27017/data_source", mongodb.GetCpuSessionNum())
	if err != nil {
		log.Error(err)
		return
	}
	defer c.Close()
	c.Debug()
	p := mongodb.MulQueryParam{}
	p.DbName = "data_source"
	p.ColName = "useraction"
	p.Query = bson.M{"timestamp": "1449417600"}
	//p.Size = 1000000
	p.Fun = func(info map[string]interface{}) {
		ad := info["AD"].(string)
		ua := info["UA"].(string)
		tag := info["tag"].([]interface{})
		for _, v := range tag {
			vm := v.(map[string]interface{})
			tid := vm["tagId"].(string)
			ldb.HSet("xxxx_"+tid, ad+"_"+ua, "1")
			ldb.HSet("xxxx_ad_"+ad+"_"+ua, tid, "1")
		}
	}
	bt := time.Now()
	fmt.Println(c.Query(p))
	fmt.Println(time.Now().Sub(bt).Seconds())

	cc, err := redis.Dial("tcp", "192.168.1.199:6380")
	if err != nil {
		fmt.Println(err)
		return
	}

	cc.Do("SELECT", "10")
	if l, err := ldb.Keys("xxxx_ad_"); err == nil {
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

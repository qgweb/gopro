//聚会算统计服务
//线上对应的链接 www.juhuisuan.com/ct/?tp=type&ne=value
package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"runtime"
	"sync"
	"time"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var (
	mhost      = flag.String("mongo-host", "192.168.1.199", "mongo地址")
	mport      = flag.String("mongo-port", "27017", "mongo端口")
	muser      = flag.String("mongo-user", "juhuisuan", "mongo用户名")
	mpwd       = flag.String("mongo-pwd", "juhuisuan", "mongo密码")
	mdbname    = flag.String("mongo-db", "jhs_tj", "mongo数据库名称")
	hhost      = flag.String("http-host", "0.0.0.0", "http服务地址")
	hport      = flag.String("http-port", "12345", "http服务端口")
	mdbsession *mgo.Session
	mux        sync.Mutex
)

func init() {
	flag.Parse()
}

func getSession() *mgo.Session {
	mux.Lock()
	defer mux.Unlock()

	if mdbsession == nil {
		var err error
		mdbsession, err = mgo.Dial(fmt.Sprintf("%s:%s@%s:%s/%s", *muser, *mpwd, *mhost, *mport, *mdbname))
		if err != nil {
			panic(err)
		}
	}
	//高并发下会关闭连接,ping下会恢复
	mdbsession.Ping()
	return mdbsession.Copy()
}

//统计
func goodsStatistics(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if msg := recover(); msg != nil {
			log.Println(msg)
		}
	}()

	//获取参数
	var (
		t = r.URL.Query().Get("tp")
		v = r.URL.Query().Get("ne")
	)

	if t == "" || v == "" {
		NotFound(w, r)
	}

	//存入mongo
	var date = time.Now().Format("2006-01-02")
	session := getSession()
	defer session.Close()

	_, err := session.DB(*mdbname).C("jhs_other_tj").Upsert(
		bson.M{"date": date, "type": t, "value": v}, bson.M{"$inc": bson.M{"count": 1}})
	if err != nil && err != io.EOF {
		log.Println(err)
	}
}

// 404 方法
func NotFound(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Location", "http://www.juhuisuan.com")
	w.WriteHeader(302)
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	http.HandleFunc("/", NotFound)
	http.HandleFunc("/gtj", goodsStatistics)
	log.Printf("程序开始启动,监听地址为%s", *hhost+":"+*hport)
	err := http.ListenAndServe(*hhost+":"+*hport, nil)
	log.Println(err)
}

package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"goclass/encrypt"

	"gopkg.in/ini.v1"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var (
	IniFile    *ini.File
	mux        sync.Mutex
	mdbsession *mgo.Session
	err        error
	conf       = flag.String("conf", "conf.ini", "配置文件")
)

func init() {
	flag.Parse()
	data, err := ioutil.ReadFile(*conf)
	if err != nil {
		log.Fatalln("读取配置文件失败,错误信息为:", err)
	}

	IniFile, err = ini.Load(data)
	if err != nil {
		log.Fatalln("加载配置文件内容失败,错误信息为:", err)
	}
}

func main() {
	var (
		outdir   = IniFile.Section("default").Key("outdir").String()
		fileName = outdir + "/" + time.Now().Add(-time.Hour*24).Format("20060102")
	)
	f, err := os.Create(fileName)
	if err != nil {
		log.Fatalln("创建文件失败")
	}
	defer f.Close()

	for k, v := range getCates() {
		res := createTjData(v)
		f.WriteString(fmt.Sprintf("%s\t%s\n", k, res))
	}
}

func createTjData(cate string) string {
	sess := GetSession()
	defer sess.Close()
	var (
		modb     = IniFile.Section("mongo").Key("db").String()
		mkey     = IniFile.Section("default").Key("mkey").String()
		clock, _ = IniFile.Section("default").Key("clock").Int()
		ndate    = time.Now().Format("2006-01-02")
		pdate    = time.Now().Add(-time.Hour * 24).Format("2006-01-02")
		aulist   = make(map[string]bool)
		cateList []string
	)

	var fun = func(d string, ct string, cl string) {
		var list []map[string]interface{}
		sess.DB(modb).C(mkey).Find(bson.M{"date": d, "cids." + cl: bson.RegEx{Pattern: ",?" + ct + ",?"}}).
			Select(bson.M{"ad": 1, "ua": 1, "_id": 0}).All(&list)

		for _, v := range list {
			ad := v["ad"].(string)
			ua := encrypt.GetEnDecoder(encrypt.TYPE_BASE64).Encode(v["ua"].(string))
			aulist[ad+"_"+ua] = true
		}
	}

	cateList = strings.Split(cate, ",")

	for i := clock; i <= 23; i++ {
		for _, v := range cateList {
			fun(pdate, v, fmt.Sprintf("%02d", i))
		}

	}
	for i := 0; i <= clock; i++ {
		for _, v := range cateList {
			fun(ndate, v, fmt.Sprintf("%02d", i))
		}
	}

	var str string
	for k, _ := range aulist {
		str += k + ","
	}

	if str != "" {
		return str[0 : len(str)-1]
	}

	return ""
}

//获取统计分类
func getCates() map[string]string {
	return IniFile.Section("cate").KeysHash()
}

//获取mongo数据库链接
func GetSession() *mgo.Session {
	mux.Lock()
	defer mux.Unlock()

	var (
		mouser = IniFile.Section("mongo").Key("user").String()
		mopwd  = IniFile.Section("mongo").Key("pwd").String()
		mohost = IniFile.Section("mongo").Key("host").String()
		moport = IniFile.Section("mongo").Key("port").String()
		modb   = IniFile.Section("mongo").Key("db").String()
	)

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

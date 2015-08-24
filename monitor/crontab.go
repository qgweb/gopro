package main

import (
	"fmt"
	"html/template"
	"net/http"
	"strings"
	"sync"
	"time"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var (
	mdbsession *mgo.Session
	mux        sync.Mutex
)

type Crontab struct {
	IP    string
	Name  string //显示名称
	Time  string
	JName string //任务名称
}

type ExecCrontabs struct {
	Cron   Crontab
	Btime  string  //开始时间
	Etime  string  //结束时间
	Status string  //状态
	Time   float64 //执行时间
}

//获取mongo-session
func GetSession() *mgo.Session {
	mux.Lock()
	defer mux.Unlock()

	if mdbsession == nil {
		var (
			err    error
			mouser = iniFile.Section("mongo").Key("user").String()
			mopwd  = iniFile.Section("mongo").Key("pwd").String()
			mohost = iniFile.Section("mongo").Key("host").String()
			moport = iniFile.Section("mongo").Key("port").String()
			modb   = iniFile.Section("mongo").Key("db").String()
		)

		mdbsession, err = mgo.Dial(fmt.Sprintf("%s:%s@%s:%s/%s", mouser, mopwd, mohost, moport, modb))
		if err != nil {
			panic(err)
		}
	}
	//高并发下会关闭连接,ping下会恢复
	mdbsession.Ping()
	return mdbsession.Copy()
}

//读取任务计划配置文件内容
func readCrontabIniData() []Crontab {
	keys := iniFile.Section("crontabjob").Keys()
	cronList := make([]Crontab, 0, 10)
	for _, v := range keys {
		ct := Crontab{}
		vvs := strings.Split(v.String(), "|")
		ct.Name = v.Name()
		ct.IP = vvs[0]
		ct.JName = vvs[1]
		ct.Time = vvs[2]

		cronList = append(cronList, ct)
	}
	return cronList
}

//读取mongo合成数据
func getData(date string) []ExecCrontabs {
	sess := GetSession()
	defer sess.Close()
	db := iniFile.Section("mongo").Key("db").String()
	crons := readCrontabIniData()
	ecrons := make([]ExecCrontabs, 0, 10)

	for _, v := range crons {
		mp := make(map[string]interface{})
		ec := ExecCrontabs{}
		ec.Cron = v
		ec.Status = "NO"

		err := sess.DB(db).C("cron").Find(bson.M{"date": date, "ip": v.IP, "name": v.JName}).
			Select(bson.M{"btime": 1, "etime": 1}).One(&mp)

		if err == mgo.ErrNotFound {
			ec.Status = "NO"
		}

		if err == nil {
			ec.Btime = mp["btime"].(string)
			ec.Etime = mp["etime"].(string)
			ec.Status = "YES"
			t1, _ := time.Parse("15:04:05", ec.Btime)
			t2, _ := time.Parse("15:04:05", ec.Etime)
			ec.Time = t2.Sub(t1).Seconds()
		}

		ecrons = append(ecrons, ec)
	}

	return ecrons
}

//任务计划
func CronJob(w http.ResponseWriter, r *http.Request) {
	date := r.URL.Query().Get("date")
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}

	t, _ := template.ParseFiles(exePath+"/tpl/cron.html")
	list := getData(date)

	t.Execute(w, list)
}

//接收任务计划
func ReceiveCronJob(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	ip := query.Get("ip")
	name := query.Get("name")
	btime := query.Get("btime")
	etime := query.Get("etime")
	date := time.Now().Format("2006-01-02")
	db := iniFile.Section("mongo").Key("db").String()

	sess := GetSession()
	defer sess.Close()

	sess.DB(db).C("cron").Upsert(bson.M{"date": date, "ip": ip, "name": name},
		bson.M{"btime": btime, "etime": etime, "date": date, "ip": ip, "name": name})
	w.Write([]byte("ok"))
}

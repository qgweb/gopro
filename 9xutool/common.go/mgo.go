package common

import (
	"fmt"
	"sync"

	"github.com/ngaut/log"
	"gopkg.in/mgo.v2"
)

type MgoConfig struct {
	UserName string
	UserPwd  string
	Host     string
	Port     string
	DBName   string
}

type MgoPool struct {
	db *mgo.Session
	sync.Mutex
	conf *MgoConfig
}

func NewMgoPool(conf *MgoConfig) *MgoPool {
	return &MgoPool{conf: conf}
}

func (this *MgoPool) Get() *mgo.Session {
	this.Lock()
	defer this.Unlock()

	var (
		url1 = fmt.Sprintf("%s:%s@%s:%s/%s", this.conf.UserName, this.conf.UserPwd,
			this.conf.Host, this.conf.Port, this.conf.DBName)
		url2 = fmt.Sprintf("%s:%s/%s", this.conf.Host, this.conf.Port, this.conf.DBName)
		url  = url1
	)

	if this.conf.UserName == "" && this.conf.UserPwd == "" {
		url = url2
	}

	if this.db == nil {
		var err error
		this.db, err = mgo.Dial(url)
		if err != nil {
			log.Fatal(err)
			return nil
		}
	}
	//高并发下会关闭连接,ping下会恢复
	this.db.Ping()
	return this.db.Copy()
}

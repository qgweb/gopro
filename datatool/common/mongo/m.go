package mongo

import (
	"fmt"
	"net"
	"sync"
	"time"

	ossh "code.google.com/p/go.crypto/ssh"
	"github.com/ngaut/log"
	"github.com/qgweb/gopro/lib/ssh"
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
	sess *mgo.Session
	sync.Mutex
}

func (this *MgoPool) getMongo(url string, f func() (net.Conn, error)) (*mgo.Session, error) {
	info, err := mgo.ParseURL(url)
	if err != nil {
		return nil, err
	}
	info.Timeout = 10 * time.Second
	info.DialServer = func(addr *mgo.ServerAddr) (net.Conn, error) {
		return f()
	}

	return mgo.DialWithInfo(info)
}

// 获取远程对象
func (this *MgoPool) GetRemoteSession(conf *MgoConfig, sshclient *ossh.Client) *mgo.Session {
	this.Lock()
	defer this.Unlock()

	url := fmt.Sprintf("%s:%s/%s", conf.Host, conf.Port, conf.DBName)
	if conf.UserName != "" && conf.UserPwd != "" {
		url = fmt.Sprintf("%s:%s@%s", conf.UserName, conf.UserPwd, url)
	}

	if this.sess == nil {
		sess, err := this.getMongo(url, func() (net.Conn, error) {
			return sshclient.Dial("tcp", fmt.Sprintf("%s:%s", conf.Host, conf.Port))
		})
		if err != nil {
			log.Fatal(err)
			return nil
		}
		this.sess = sess
	}

	this.sess.Ping()
	return this.sess.Clone()
}

// 获取本地session对象
func (this *MgoPool) GetLocalSession(conf *MgoConfig) *mgo.Session {
	this.Lock()
	defer this.Unlock()

	url := fmt.Sprintf("%s:%s/%s", conf.Host, conf.Port, conf.DBName)
	if conf.UserName != "" && conf.UserPwd != "" {
		url = fmt.Sprintf("%s:%s@%s", conf.UserName, conf.UserPwd, url)
	}

	if this.sess == nil {
		sess, err := mgo.Dial(url)
		if err != nil {
			log.Fatal(err)
			return nil
		}
		this.sess = sess
	}

	this.sess.Ping()
	return this.sess.Clone()
}

// 获取ssh对象
func GetSSHLinker(conf ssh.Config) (*ssh.SSHLinker, error) {
	return ssh.NewSSHLinker(conf)
}

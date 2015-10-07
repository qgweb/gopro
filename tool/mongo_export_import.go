// mongo export import tool
package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"math"
	"net"
	"sync"
	"time"

	ossh "code.google.com/p/go.crypto/ssh"
	"github.com/ngaut/log"
	"github.com/qgweb/gopro/lib/ssh"
	"gopkg.in/ini.v1"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var (
	conf        = flag.String("conf", "", "配置文件")
	remote_host = flag.String("remote_host", "122.225.98.69", "宿主机")
	remote_user = flag.String("remote_user", "root", "宿主机用户")
	remote_key  = flag.String("remote_key", "", "本机的ssh私钥")

	remote_mongo_host  = flag.String("remote_mongo_host", "", "远程mongo_host")
	remote_mongo_port  = flag.String("remote_mongo_port", "", "远程mongo_port")
	remote_mongo_db    = flag.String("remote_mongo_db", "", "远程mongo_db")
	remote_mongo_user  = flag.String("remote_mongo_user", "", "远程mongo_user")
	remote_mongo_pwd   = flag.String("remote_mongo_pwd", "", "远程mongo_pwd")
	remote_mongo_table = flag.String("remote_mongo_table", "", "远程mongo_collection")

	local_mongo_host  = flag.String("local_mongo_host", "", "本地mongo_host")
	local_mongo_port  = flag.String("local_mongo_port", "", "本地mongo_port")
	local_mongo_db    = flag.String("local_mongo_db", "", "本地mongo_db")
	local_mongo_user  = flag.String("local_mongo_user", "", "本地mongo_user")
	local_mongo_pwd   = flag.String("local_mongo_pwd", "", "本地mongo_pwd")
	local_mongo_table = flag.String("local_mongo_table", "", "本地mongo_collection")

	to = flag.String("to", "1", "数据导入方向，1：远程到本地，2：本地到远程")

	iniFile *ini.File
	err     error
	linker  *ssh.SSHLinker
)

func init() {
	flag.Parse()

	//判断是否有配置文件
	if *conf != "" {
		data, err := ioutil.ReadFile(*conf)
		if err != nil {
			log.Fatal(err)
			return
		}

		iniFile, err = ini.Load(data)
		if err != nil {
			log.Fatal(err)
			return
		}

		//读取数据
		*to = iniFile.Section("default").Key("to").String()

		*remote_host = iniFile.Section("remote").Key("remote_host").String()
		*remote_user = iniFile.Section("remote").Key("remote_user").String()
		*remote_key = iniFile.Section("remote").Key("remote_key").String()

		*remote_mongo_host = iniFile.Section("remote-mongo").Key("remote_mongo_host").String()
		*remote_mongo_port = iniFile.Section("remote-mongo").Key("remote_mongo_port").String()
		*remote_mongo_db = iniFile.Section("remote-mongo").Key("remote_mongo_db").String()
		*remote_mongo_user = iniFile.Section("remote-mongo").Key("remote_mongo_user").String()
		*remote_mongo_pwd = iniFile.Section("remote-mongo").Key("remote_mongo_pwd").String()
		*remote_mongo_table = iniFile.Section("remote-mongo").Key("remote_mongo_table").String()

		*local_mongo_host = iniFile.Section("local-mongo").Key("local_mongo_host").String()
		*local_mongo_port = iniFile.Section("local-mongo").Key("local_mongo_port").String()
		*local_mongo_db = iniFile.Section("local-mongo").Key("local_mongo_db").String()
		*local_mongo_user = iniFile.Section("local-mongo").Key("local_mongo_user").String()
		*local_mongo_pwd = iniFile.Section("local-mongo").Key("local_mongo_pwd").String()
		*local_mongo_table = iniFile.Section("local-mongo").Key("local_mongo_table").String()
	}

	// 初始化ssh linker
	ssh_conf := ssh.Config{}
	ssh_conf.PrivaryKey = ssh.GetPrivateKey()
	ssh_conf.RemoteHost = *remote_host
	ssh_conf.RemotePort = 22
	ssh_conf.RemoteUser = *remote_user
	linker, err = ssh.NewSSHLinker(ssh_conf)
	if err != nil {
		log.Fatal(err)
		return
	}
}

func getMongo(url string, f func() (net.Conn, error)) (*mgo.Session, error) {
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

func getRemoteSession(client *ossh.Client) *mgo.Session {
	url := ""
	if *remote_mongo_user == "" && *remote_mongo_pwd == "" {
		url = fmt.Sprintf("%s:%s/%s", *remote_mongo_host, *remote_mongo_port, *remote_mongo_db)
	} else {
		url = fmt.Sprintf("%s:%s@%s:%s/%s", *remote_mongo_user, *remote_mongo_pwd,
			*remote_mongo_host, *remote_mongo_port, *remote_mongo_db)
	}

	sess, err := getMongo(url, func() (net.Conn, error) {
		return client.Dial("tcp", fmt.Sprintf("%s:%s", *remote_mongo_host, *remote_mongo_port))
	})

	if err != nil {
		log.Fatal(err)
		return nil
	}

	return sess
}

func getLocalSesssion() *mgo.Session {
	url := ""
	if *local_mongo_user == "" && *local_mongo_pwd == "" {
		url = fmt.Sprintf("%s:%s/%s", *local_mongo_host, *local_mongo_port, *local_mongo_db)
	} else {
		url = fmt.Sprintf("%s:%s@%s:%s/%s", *local_mongo_user, *local_mongo_pwd,
			*local_mongo_host, *local_mongo_port, *local_mongo_db)
	}
	mdbsession, err := mgo.Dial(url)
	if err != nil {
		log.Fatal(err)
		return nil
	}
	return mdbsession
}

func remote_to_local(rsession *mgo.Session, lsession *mgo.Session) {
	var (
		pageSize  = 10000
		page      = 1
		pageTotal = 0
		wg        = sync.WaitGroup{}
	)

	count, err := rsession.DB(*remote_mongo_db).C(*remote_mongo_table).Find(bson.M{}).Count()
	if err != nil {
		log.Fatal(err)
	}
	pageTotal = int(math.Ceil(float64(count) / float64(pageSize)))
	for ; page <= pageTotal; page++ {
		wg.Add(1)
		go func(i int) {
			var list []bson.M
			var rs = rsession.Clone()
			var ls = lsession.Clone()

			err = rs.DB(*remote_mongo_db).C(*remote_mongo_table).Find(bson.M{}).
				Skip((i - 1) * pageSize).Limit(pageSize).All(&list)
			if err == nil && len(list) > 0 {
				tmp := make([]interface{}, 0, len(list))
				for _, v := range list {
					log.Info(v)
					tmp = append(tmp, v)
				}

				ls.DB(*local_mongo_db).C(*local_mongo_table).Insert(tmp...)
			}
			rs.Close()
			ls.Close()
			wg.Done()
		}(page)
	}
	wg.Wait()
	rsession.Close()
	lsession.Close()
}

func local_to_remote(rsession *mgo.Session, lsession *mgo.Session) {
	var (
		pageSize  = 10000
		page      = 1
		pageTotal = 0
		wg        = sync.WaitGroup{}
	)

	count, err := lsession.DB(*local_mongo_db).C(*local_mongo_table).Find(bson.M{}).Count()
	if err != nil {
		log.Fatal(err)
	}
	pageTotal = int(math.Ceil(float64(count) / float64(pageSize)))
	for ; page <= pageTotal; page++ {
		wg.Add(1)
		go func(i int) {
			var list []bson.M
			var rs = rsession.Clone()
			var ls = lsession.Clone()

			err = ls.DB(*local_mongo_db).C(*local_mongo_table).Find(bson.M{}).
				Skip((i - 1) * pageSize).Limit(pageSize).All(&list)
			if err == nil && len(list) > 0 {
				tmp := make([]interface{}, 0, len(list))
				for _, v := range list {
					log.Info(v)
					tmp = append(tmp, v)
				}

				rs.DB(*remote_mongo_db).C(*remote_mongo_table).Insert(tmp...)
			}
			rs.Close()
			ls.Close()
			wg.Done()
		}(page)
	}
	wg.Wait()
	rsession.Close()
	lsession.Close()
}

func main() {
	client := linker.GetClient()
	defer linker.Close()
	if *to == "1" {
		remote_to_local(getRemoteSession(client), getLocalSesssion())
	}
	if *to == "2" {
		local_to_remote(getRemoteSession(client), getLocalSesssion())
	}
}

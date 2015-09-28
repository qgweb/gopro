// grab taocat
package main

import (
	"flag"
	"fmt"
	"github.com/ngaut/log"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var (
	mdbsession *mgo.Session
	mo_user    = "xu"
	mo_pwd     = "123456"
	mo_host    = "192.168.1.199"
	mo_port    = "27017"
	mo_db      = "xu_precise"
	mo_table   = "tao_cat"
)

type Category struct {
	Name  string     `bson:"name"`
	Spell string     `bson:"spell"`
	Sid   string     `bson:"cid"`
	Id    string     `bson:"id"` //顶级需要
	Level int        `bson:"level"`
	Child []Category `bson:"child"` //子集
	Pid   string     `bson:"pid"`
	Type  string     `bson:"type"`
}

//获取mongo数据库链接
func GetSession() *mgo.Session {
	var (
		mouser = mo_user
		mopwd  = mo_pwd
		mohost = mo_host
		moport = mo_port
		modb   = mo_db
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

// 活取所有分类记录
func getAllCatroy() []Category {
	sess := GetSession()
	defer sess.Close()
	var list []Category
	sess.DB("xu_precise").C(mo_table).Find(bson.M{}).All(&list)
	return list
}

// 生成树
func catTree(data []Category, pid string) []Category {
	list := make([]Category, 0, 1000)
	for _, v := range data {
		if v.Pid == pid {
			v.Child = catTree(data, v.Sid)
			list = append(list, v)
		}
	}
	return list
}

// 查询某个节点
func queryCategory(list []Category, sid string) []Category {
	for _, v := range list {
		if v.Sid == sid {
			return append([]Category{}, v)
		} else {
			if len(v.Child) > 0 {
				if vv := queryCategory(v.Child, sid); len(vv) > 0 {
					return vv
				}
			}
		}
	}
	return nil
}

// 输出树形结构
func display(data []Category, level int, prefix string) {
	for _, v := range data {
		if v.Level > level {
			break
		}

		log.Info(prefix+v.Name, "", v.Sid)
		if len(v.Child) > 0 {
			display(v.Child, level, "---"+prefix)
		}
	}
}

func main() {
	var (
		sid   = flag.String("sid", "", "分类id")
		level = flag.Int("level", 5, "等级")
	)

	flag.Parse()

	log.SetHighlighting(true)
	list := catTree(getAllCatroy(), "")

	if *sid != "" {
		display(queryCategory(list, *sid), *level, "")
	} else {
		display(list, *level, "")
	}

	//log.Error(SecondLevelCategory("next", "50011665", 3))
}

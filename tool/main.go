package main

import (
	"fmt"

	"github.com/qgweb/gopro/lib/ssh"
	"gopkg.in/mgo.v2"
	//"gopkg.in/mgo.v2/bson"
	"html/template"
	"io"
	"log"
	"math/rand"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/qgweb/new/lib/convert"
	"gopkg.in/mgo.v2/bson"
)

var sess *mgo.Session

func getSession() *mgo.Session {
	if sess == nil {
		s := ssh.Config{}
		s.PrivaryKey = ssh.GetPrivateKey()
		s.RemoteHost = "122.225.98.69"
		s.RemotePort = 22
		s.RemoteUser = "root"
		sh, err := ssh.NewSSHLinker(s)
		if err != nil {
			log.Fatalln(err)
		}
		di, err := mgo.ParseURL("mongodb://192.168.0.68:10003")
		di.DialServer = func(addr *mgo.ServerAddr) (net.Conn, error) {
			return sh.GetClient().Dial("tcp", "192.168.0.68:10003")
		}
		ss, err := mgo.DialWithInfo(di)
		if err != nil {
			log.Fatalln(err)
		}
		sess = ss
	}
	return sess.Clone()
}

type TBTag struct {
	ObjId string  `json:"oid"`
	Name  string  `json:"name"`
	Count string  `json:"count"`
	Cid   string  `json:"cid"`
	Pid   string  `json:"pid"`
	Prfix string  `json:"pre"`
	Child []TBTag `json:"child"`
}

func createTaocat(list []TBTag, pid string, prefix string) []TBTag {
	var nlist = make([]TBTag, 0, 10)
	for _, v := range list {
		if v.Pid == pid {
			v.Child = createTaocat(list, v.Cid, prefix+"-")
			v.Prfix = prefix
			nlist = append(nlist, v)
		}
	}
	return nlist
}

func TaoCat(w http.ResponseWriter, r *http.Request) {
	var list = make([]TBTag, 0, 50)
	sess := getSession()
	iter := sess.DB("xu_precise").C("taocat_big").Find(nil).Iter()
	for {
		var a map[string]interface{}
		if !iter.Next(&a) {
			break
		}
		tb := TBTag{}
		tb.Cid = convert.ToString(a["cid"])
		tb.Pid = convert.ToString(a["pid"])
		tb.Name = convert.ToString(a["name"])
		tb.Count = convert.ToString(a["count"])
		tb.ObjId = a["_id"].(bson.ObjectId).Hex()
		list = append(list, tb)
	}
	sess.Close()
	nlist := createTaocat(list, "", "-")

	t, err := template.ParseFiles("tc.html")
	fmt.Println(err)
	t.Execute(w, nlist)
}

func SaveTb(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		var cid = strings.TrimSpace(r.PostFormValue("cid"))
		var val = strings.TrimSpace(r.PostFormValue("val"))
		if cid == "" || val == "" {
			io.WriteString(w, "有数值为空！！！")
			return
		}

		sess := getSession()
		err := sess.DB("xu_precise").C("taocat_big").Update(bson.M{"_id": bson.ObjectIdHex(cid)}, bson.M{"$set": bson.M{"count": val}})
		if err == nil {
			io.WriteString(w, "修改成功")
		} else {
			io.WriteString(w, "修改失败")
		}

		sess.Close()
	}
}

func createUrlTag(list []TBTag, pid string, prefix string) []TBTag {
	var nlist = make([]TBTag, 0, 10)
	for _, v := range list {
		if v.Pid == pid {
			v.Child = createTaocat(list, v.ObjId, prefix+"-")
			v.Prfix = prefix
			nlist = append(nlist, v)
		}
	}
	return nlist
}
func UrlCat(w http.ResponseWriter, r *http.Request) {
	var list = make([]TBTag, 0, 50)
	sess := getSession()
	iter := sess.DB("precise").C("domain_category").Find(nil).Iter()
	for {
		var a map[string]interface{}
		if !iter.Next(&a) {
			break
		}
		tb := TBTag{}
		tb.Cid = convert.ToString(a["cid"])
		if v, ok := a["pid"].(bson.ObjectId); ok {
			tb.Pid = v.Hex()
		} else {
			tb.Pid = ""
		}
		tb.Name = convert.ToString(a["name"])
		tb.Count = convert.ToString(a["count"])
		tb.ObjId = a["_id"].(bson.ObjectId).Hex()
		list = append(list, tb)
	}
	sess.Close()
	//updateDomain(list)
	nlist := createUrlTag(list, "", "-")

	t, err := template.ParseFiles("url.html")
	fmt.Println(err)
	t.Execute(w, nlist)
}

func SaveUrl(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		var cid = strings.TrimSpace(r.PostFormValue("cid"))
		var val = strings.TrimSpace(r.PostFormValue("val"))
		if cid == "" || val == "" {
			io.WriteString(w, "有数值为空！！！")
			return
		}

		sess := getSession()
		err := sess.DB("precise").C("domain_category").Update(bson.M{"_id": bson.ObjectIdHex(cid)}, bson.M{"$set": bson.M{"count": val}})
		if err == nil {
			io.WriteString(w, "修改成功")
		} else {
			io.WriteString(w, "修改失败")
		}

		sess.Close()
	}
}

func getRandNum(bnum, enum int) int {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return r.Intn(enum-bnum) + bnum
}

func updateDomain(list []TBTag) {
	sess := getSession()
	for _, v := range list {
		num := getRandNum(1000, 20000)
		sess.DB("precise").C("domain_category").Update(bson.M{"_id": bson.ObjectIdHex(v.ObjId)}, bson.M{"$set": bson.M{"count": num}})
	}
	sess.Close()
}

func main() {
	http.HandleFunc("/tc", TaoCat)
	http.HandleFunc("/stc", SaveTb)
	http.HandleFunc("/url", UrlCat)
	http.HandleFunc("/surl", SaveUrl)
	log.Fatalln(http.ListenAndServe(":8080", nil))
}

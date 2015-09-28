// grab taocat
package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/bitly/go-simplejson"
	iconv "github.com/djimenez/iconv-go"
	"github.com/ngaut/log"
	"github.com/qgweb/gopro/lib/convert"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const (
	TAO_CAT_URL = "https://upload.taobao.com/auction/json/reload_cats.htm?customId="
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

// 获取html
func grabHtml(path string, sid string) string {
	v := url.Values{}
	v.Set("path", path)
	v.Set("sid", sid)
	v.Set("pv", "")
	body := strings.NewReader(v.Encode()) //把form数据编下码

	client := &http.Client{}
	r, _ := http.NewRequest("POST", TAO_CAT_URL, body)

	r.Header.Set("Accept", "application/json, text/javascript, */*; q=0.01")
	r.Header.Set("Accept-Language", "zh-CN,zh;q=0.8")
	r.Header.Set("Cache-Control", "no-cache")
	r.Header.Set("Connection", "keep-alive")
	r.Header.Set("Content-Length", "17")
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	r.Header.Set("Cookie", "cna=iBpHDuxbpDoCAXPI8O66zMhi; thw=cn; ali_ab=122.235.237.26.1438678183372.3; miid=6856742623729070531; x=e%3D1%26p%3D*%26s%3D0%26c%3D0%26f%3D0%26g%3D0%26t%3D0%26__ll%3D-1%26_ato%3D0; v=0; _tb_token_=PsM1KPAjxqah5f6; isg=31DEBB4066877FAAB35DDA6EB47D4AAA; uc3=nk2=F4T%2BqCs1GCv3cUU%3D&id2=UU6p%2B6jaIpgS&vt3=F8dASMlJs8zUtnhkm80%3D&lg2=Vq8l%2BKCLz3%2F65A%3D%3D; existShop=MTQ0MzQzODI0OA%3D%3D; lgc=tracyxiang5; tracknick=tracyxiang5; sg=57c; cookie2=1c47414352a167e02fb5dd9058a5a33c; mt=np=&ci=9_1&cyk=0_0; cookie1=ACiySN0X98ZST2xXkglRaGFbZmpUopo7AQbwVps1wd8%3D; unb=268142207; skt=f779130c1598812d; t=afd38437622a33dd1817869da512ff1e; publishItemObj=Ng%3D%3D; _cc_=U%2BGCWk%2F7og%3D%3D; tg=0; _l_g_=Ug%3D%3D; _nk_=tracyxiang5; cookie17=UU6p%2B6jaIpgS; l=AmBg3efaZReJJcinimEK4qBisGAycEQz; uc1=cookie14=UoWzWiy1dE99ww%3D%3D&existShop=true&cookie16=VT5L2FSpNgq6fDudInPRgavC%2BQ%3D%3D&cookie21=W5iHLLyFfXVRDP8mxoRA8A%3D%3D&tag=3&cookie15=VT5L2FSpMGV7TQ%3D%3D&pas=0")
	r.Header.Set("Host", "upload.taobao.com")
	r.Header.Set("Pragma", "no-cache")
	r.Header.Set("Referer", "http://upload.taobao.com/auction/sell.jhtml?spm=a1z0e.1.0.0.nigjAo&mytmenu=wym&utkn=g,orzgcy3zpbuwc3thgu1435198089453&scm=1028.1.1.101")
	r.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/42.0.2311.152 Safari/537.36")
	r.Header.Set("X-Requested-With", "XMLHttpRequest")

	resp, err := client.Do(r)
	if err != nil {
		if resp == nil {
			panic(err)
		}
	}

	defer resp.Body.Close()

	a, _ := ioutil.ReadAll(resp.Body)
	d, _ := iconv.ConvertString(string(a), "gbk", "utf-8")
	return d
}

// 获取第一级类目(包含第二)
func FirstLevelCategory(path string, sid string) []Category {
	data := grabHtml(path, sid)
	jn, err := simplejson.NewJson([]byte(data))
	if err != nil {
		log.Fatal("cookie失效")
		return nil
	}

	if v, _ := jn.GetIndex(0).Get("pName").String(); v == "类目" {
		//顶级类目
		firstList, _ := jn.GetIndex(0).Get("data").Array()
		fcs := make([]Category, 0, 100)
		for _, fdata := range firstList {
			fdataJson := simplejson.New()
			fdataJson.SetPath([]string{}, fdata)
			fc := Category{}
			fc.Name, _ = fdataJson.Get("name").String()
			id, _ := fdataJson.Get("id").Int()
			fc.Sid = convert.ToString(id)
			fc.Level = 1
			fc.Child = make([]Category, 0, 100)

			//二级类目
			secondList, _ := fdataJson.Get("data").Array()
			for _, sdata := range secondList {
				sdataJson := simplejson.New()
				sdataJson.SetPath([]string{}, sdata)
				sfc := Category{}
				sfc.Name, _ = sdataJson.Get("name").String()
				sfc.Sid, _ = sdataJson.Get("sid").String()
				sfc.Level = 2
				sfc.Spell, _ = sdataJson.Get("spell").String()
				fc.Child = append(fc.Child, sfc)
			}

			log.Info(fc.Name, "\t", fc.Id)
			for _, v := range fc.Child {
				log.Info("-", v.Name, "\t", v.Sid, "\t", v.Spell)
			}
			fcs = append(fcs, fc)
		}
		return fcs
	}
	return nil
}

// 递归分类
func SecondLevelCategory(path string, sid string, level int) []Category {
	time.Sleep(time.Millisecond * 100)
	data := grabHtml(path, sid)
	jn, err := simplejson.NewJson([]byte(data))
	if err != nil {
		log.Error("cookie失效", err)
		return nil
	}

	if v, _ := jn.GetIndex(0).Get("pName").String(); v == "类目" {
		firstList, _ := jn.GetIndex(0).Get("data").Array()
		fcs := make([]Category, 0, 100)
		for _, fdata := range firstList {
			fdataJson := simplejson.New()
			fdataJson.SetPath([]string{}, fdata)
			secondlist, _ := fdataJson.Get("data").Array()
			for _, sdata := range secondlist {
				sdataJson := simplejson.New()
				sdataJson.SetPath([]string{}, sdata)
				fc := Category{}
				fc.Name, _ = sdataJson.Get("name").String()
				fc.Sid, _ = sdataJson.Get("sid").String()
				fc.Spell, _ = sdataJson.Get("spell").String()
				fc.Level = level
				fc.Type = "0"
				log.Info(fc)
				fc.Child = SecondLevelCategory(path, fc.Sid, level+1)
				fcs = append(fcs, fc)
			}
		}
		return fcs
	}

	if v, _ := jn.GetIndex(0).Get("pName").String(); strings.Contains(v, "品牌") {
		firstList, _ := jn.GetIndex(0).Get("data").Array()
		fcs := make([]Category, 0, 100)
		for _, fdata := range firstList {
			fdataJson := simplejson.New()
			fdataJson.SetPath([]string{}, fdata)
			secondlist, _ := fdataJson.Get("data").Array()
			for _, sdata := range secondlist {
				sdataJson := simplejson.New()
				sdataJson.SetPath([]string{}, sdata)
				fc := Category{}
				fc.Name, _ = sdataJson.Get("name").String()
				fc.Sid, _ = sdataJson.Get("sid").String()
				fc.Spell, _ = sdataJson.Get("spell").String()
				fc.Level = level
				fc.Type = "1"
				log.Warn(fc)
				//fc.Child = SecondLevelCategory(path, fc.Sid, level+1)
				fcs = append(fcs, fc)
			}
		}
		return fcs
	}

	return nil
}

// 添加monggo记录
func addRecord(list []Category, pid string) {
	sess := GetSession()
	defer sess.Close()

	for _, v := range list {
		sess.DB("xu_precise").C(mo_table).Upsert(bson.M{"cid": v.Sid}, bson.M{
			"$set": bson.M{
				"name":  v.Name,
				"spell": v.Spell,
				"level": v.Level,
				"pid":   pid,
				"type":  v.Type,
			},
		})
		if len(v.Child) > 0 {
			addRecord(v.Child, v.Sid)
		}
	}
}

func main() {
	var (
		pid = flag.String("pid", "", "父节点id")
	)

	flag.Parse()

	if *pid == "" {
		log.Fatal("木有父节点")
	}

	log.SetHighlighting(true)
	// list := FirstLevelCategory("all", "")
	// addRecord(list, "")
	//log.Info(list)
	list := SecondLevelCategory("next", *pid, 3)
	addRecord(list, *pid)
}

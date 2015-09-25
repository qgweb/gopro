// grab taocat
package main

import (
	"flag"
	"fmt"
	"github.com/bitly/go-simplejson"
	"github.com/ngaut/log"
	"github.com/qgweb/gopro/lib/convert"
	"github.com/qiniu/iconv"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	TAO_CAT_URL = "https://upload.taobao.com/auction/json/reload_cats.htm?customId="
)

var (
	mdbsession *mgo.Session
)

type Category struct {
	Name  string     `bson:"name"`
	Spell string     `bson:"spell"`
	Sid   string     `bson:"cid"`
	Id    string     `bson:"id"` //顶级需要
	Level int        `bson:"level"`
	Child []Category `bson:"child"` //子集
	Pid   string     `bson:"pid"`
}

//获取mongo数据库链接
func GetSession() *mgo.Session {
	var (
		mouser = "xu"
		mopwd  = "123456"
		mohost = "127.0.0.1"
		moport = "27017"
		modb   = "xu_precise"
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
	cd, _ := iconv.Open("utf-8", "gbk")
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
	r.Header.Set("Cookie", "thw=cn; cna=BLJwDlhUGU4CAX14n0a8en7J; v=0; _tb_token_=tNI3NeXl1OXO9G; uc3=nk2=F4T%2BqCs1GCv3cUU%3D&id2=UU6p%2B6jaIpgS&vt3=F8dASMr55qi6XxKWJuU%3D&lg2=UIHiLt3xD8xYTw%3D%3D; existShop=MTQ0MzE5ODY2OQ%3D%3D; lgc=tracyxiang5; tracknick=tracyxiang5; sg=57c; cookie2=1c7a14d526ade08e9628789b95114d34; mt=np=&ci=1_1; cookie1=ACiySN0X98ZST2xXkglRaGFbZmpUopo7AQbwVps1wd8%3D; unb=268142207; skt=e1eaf69f0c4f438b; t=1bb9fbc5f55889dff929db1dc7393e01; publishItemObj=Ng%3D%3D; _cc_=U%2BGCWk%2F7og%3D%3D; tg=0; _l_g_=Ug%3D%3D; _nk_=tracyxiang5; cookie17=UU6p%2B6jaIpgS; isg=F0AE4D906AC26DEDD38109B2D9759EE9; uc1=cookie14=UoWzWim8vS%2BMzg%3D%3D&existShop=true&cookie16=Vq8l%2BKCLySLZMFWHxqs8fwqnEw%3D%3D&cookie21=UIHiLt3xSixwG45%2Bs3wzsA%3D%3D&tag=3&cookie15=UIHiLt3xD8xYTw%3D%3D&pas=0; l=Ari41eRdRiqCBiCfohliZiAlCGhKGxyr")
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
	return cd.ConvString(string(a))
}

// 获取第一级类目(包含第二)
func FirstLevelCategory(path string, sid string) []Category {
	data := grabHtml(path, sid)
	jn, err := simplejson.NewJson([]byte(data))
	if err != nil {
		log.Error("cookie失效")
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
		log.Fatal("cookie失效")
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
				log.Info(fc)
				fc.Child = SecondLevelCategory(path, fc.Sid, level+1)
				fcs = append(fcs, fc)
			}
		}
		return fcs
	}

	if v, _ := jn.GetIndex(0).Get("pName").String(); v == "品牌" {

	}
	return nil
}

// 添加monggo记录
func addRecord(list []Category, pid string) {
	sess := GetSession()
	defer sess.Close()

	for _, v := range list {
		sess.DB("xu_precise").C("taocat").Upsert(bson.M{"cid": v.Sid}, bson.M{
			"$set": bson.M{
				"name":  v.Name,
				"spell": v.Spell,
				"level": v.Level,
				"pid":   pid,
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
	//list := FirstLevelCategory("all", "")

	list := SecondLevelCategory("next", *pid, 3)
	addRecord(list, *pid)
}

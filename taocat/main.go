// grab taocat
package main

import (
	"github.com/bitly/go-simplejson"
	"github.com/ngaut/log"
	"github.com/qiniu/iconv"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	TAO_CAT_URL = "https://upload.taobao.com/auction/json/reload_cats.htm?customId="
)

type Category struct {
	Name  string     `json:"name"`
	Spell string     `json:"spell"`
	Sid   string     `json:"sid"`
	Id    string     `json:"id"` //顶级需要
	Level int        `json:"level"`
	Child []Category `json:"child"` //子集
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
	r.Header.Set("Cookie", "cna=iBpHDuxbpDoCAXPI8O66zMhi; thw=cn; ali_ab=122.235.237.26.1438678183372.3; miid=6856742623729070531; x=e%3D1%26p%3D*%26s%3D0%26c%3D0%26f%3D0%26g%3D0%26t%3D0%26__ll%3D-1%26_ato%3D0; v=0; _tb_token_=PsM1KPAjxqah5f6; isg=30DBEEC6A07BEC83ECD1E9E2516C7CDC; uc3=nk2=F4T%2BqCs1GCv3cUU%3D&id2=UU6p%2B6jaIpgS&vt3=F8dASMr56U95cn4JdGw%3D&lg2=URm48syIIVrSKA%3D%3D; existShop=MTQ0MzE3NTQ4Mg%3D%3D; lgc=tracyxiang5; tracknick=tracyxiang5; sg=57c; cookie2=1c47414352a167e02fb5dd9058a5a33c; mt=np=&ci=9_1&cyk=0_0; cookie1=ACiySN0X98ZST2xXkglRaGFbZmpUopo7AQbwVps1wd8%3D; unb=268142207; skt=ac21cf765ab75d65; t=afd38437622a33dd1817869da512ff1e; publishItemObj=Ng%3D%3D; _cc_=VT5L2FSpdA%3D%3D; tg=0; _l_g_=Ug%3D%3D; _nk_=tracyxiang5; cookie17=UU6p%2B6jaIpgS; l=Ajs7zObGrvYmzIP2HSjR712GSxGlnk-S; uc1=cookie14=UoWzWimyiWEu0A%3D%3D&existShop=true&cookie16=UIHiLt3xCS3yM2h4eKHS9lpEOw%3D%3D&cookie21=UIHiLt3xSixwG45%2Bs3wzsA%3D%3D&tag=3&cookie15=U%2BGCWk%2F75gdr5Q%3D%3D&pas=0")
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
func FirstLevelCategory(path string, sid string) {
	data := grabHtml(path, sid)
	jn, err := simplejson.NewJson([]byte(data))
	if err != nil {
		log.Error("cookie失效")
		return
	}

	if v, _ := jn.GetIndex(0).Get("pName").String(); v == "类目" {
		//顶级类目
		firstList, _ := jn.GetIndex(0).Get("data").Array()
		for _, fdata := range firstList {
			fdataJson := simplejson.New()
			fdataJson.SetPath([]string{}, fdata)
			fc := Category{}
			fc.Name, _ = fdataJson.Get("name").String()
			fc.Id, _ = fdataJson.Get("id").String()
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
		}
	}
}

// 递归分类
func SecondLevelCategory(path string, sid string, level int) []Category {
	time.Sleep(time.Millisecond * 100)
	data := grabHtml(path, sid)
	jn, err := simplejson.NewJson([]byte(data))
	if err != nil {
		log.Error("cookie失效")
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

func main() {
	log.SetHighlighting(false)
	//FirstLevelCategory("all", "")
	log.Error(SecondLevelCategory("next", "50011665", 3))
}

package grab

import (
	"github.com/PuerkitoBio/goquery"
	"github.com/henrylee2cn/surfer/agent"
	"github.com/ngaut/log"
	"math/rand"
	"net/http"
	"regexp"
	"strings"
	"time"
)

//抓取京东商品页面
func GrabJDHTML(url string) string {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Error(err)
		return ""
	}

	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.8")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Pragma", "no-cache")

	if req.UserAgent() == "" {
		l := len(agent.UserAgents["common"])
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		req.Header.Set("User-Agent", agent.UserAgents["common"][r.Intn(l)])
	}
	//	client.Lock()
	//	defer client.Unlock()
	resp, err := client.Do(req)
	if err != nil {
		log.Error(err)
		return ""
	}

	if resp.Body != nil {
		defer resp.Body.Close()
		return changeCharsetEncodingAuto(resp.Body, resp.Header.Get("Content-Type"))
	}
	return ""
}

func GetJDTitle(p *goquery.Document) string {
	return strings.TrimSpace(p.Find("#itemInfo").Find("h1").Text())
}

func GetJDCategory(p *goquery.Document) []string {
	res := make([]string, 2, 2)
	l := p.Find(".breadcrumb a").Length()
	n := 3
	if l >= 3 {
		LABEL:
		obj := p.Find(".breadcrumb a").Eq(l - n)
		href, _ := obj.Attr("href")
		r, _ := regexp.Compile(`(\d+)\,(\d+)\,(\d+)`)
		mstr := r.FindStringSubmatch(href)
		res[0] = strings.TrimSpace(obj.Text())
		if len(mstr) >= 4 {
			res[1] = mstr[len(mstr)-1]
		} else {
			n = 2
			goto LABEL
		}
	}
	return res
}

func GetJDBrand(p *goquery.Document) string {
	obj:=p.Find(".breadcrumb a").Last().Prev()
	href,_ := obj.Attr("href")
	r, _ := regexp.Compile(`(\d+)\-(\d+)\.html`)
	mstr := r.FindStringSubmatch(href)
	if len(mstr) > 0 {
		return strings.TrimSpace(obj.Text())
	}
	return ""
}

func GetJDAttributes(p *goquery.Document) []string{
	elems := make([]string,0,20)
	p.Find("#parameter2 li").Each(func(index int, element *goquery.Selection) {
		elem := strings.Split(element.Text(), "：")
		if len(elem) < 2 {
			elem = strings.Split(element.Text(), ":")
		}
		if len(elem) >=2 {
			elem[0] = strings.TrimSpace(elem[0])
			elem[1] = strings.TrimSpace(elem[1])
			elems = append(elems, strings.Join(elem,":"))
		}

	})
	return elems
}
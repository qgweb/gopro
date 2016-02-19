package grab

import (
	"crypto/tls"
	"io"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"golang.org/x/net/html/charset"
	"github.com/PuerkitoBio/goquery"
	"github.com/henrylee2cn/surfer/agent"
	"github.com/ngaut/log"
	"sync"
)

var (
	client *http.Client
	mux sync.Mutex
)

func init() {
	client = buildClient()
}

func buildClient() *http.Client {
	client := &http.Client{
		CheckRedirect: func(req *http.Request, _ []*http.Request) error {
			return nil
		},
	}

	transport := &http.Transport{
		Dial: func(network, addr string) (net.Conn, error) {
			c, err := net.DialTimeout(network, addr, time.Second*30)
			if err != nil {
				log.Error(err)
				return nil, err
			}
			c.SetDeadline(time.Now().Add(time.Second * 30))
			return c, nil
		},
	}

	transport.TLSClientConfig = &tls.Config{RootCAs: nil, InsecureSkipVerify: true}
	transport.DisableCompression = true
	client.Transport = transport
	return client
}

func changeCharsetEncodingAuto(sor io.ReadCloser, contentTypeStr string) string {
	var err error
	destReader, err := charset.NewReader(sor, contentTypeStr)

	if err != nil {
		log.Error(err)
		destReader = sor
	}

	var sorbody []byte
	if sorbody, err = ioutil.ReadAll(destReader); err != nil {
		log.Error(err)
	}

	bodystr := string(sorbody)

	return bodystr
}

//解析页面到node
func ParseNode(h string) (*goquery.Document, error) {
	return goquery.NewDocumentFromReader(strings.NewReader(h))
}

//抓取淘宝商品页面
func GrabTaoHTML(url string) string {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Error(err)
		return ""
	}

	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.8")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Host", "www.taobao.com")
	req.Header.Set("Pragma", "no-cache")

	if req.UserAgent() == "" {
		l := len(agent.UserAgents["common"])
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		req.Header.Set("User-Agent", agent.UserAgents["common"][r.Intn(l)])
	}
	//	client.Lock()
	//	defer client.Unlock()
	mux.Lock()
	resp, err := client.Do(req)
	mux.Unlock()
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

//获取淘宝标题
func GetTitle(p *goquery.Document) string {
	return strings.TrimSpace(p.Find("title").Text())
}

//获取淘宝属性信息
func GetAttrbuites(p *goquery.Document) string {
	attribute := make([]string, 0, 20)
	p.Find("#J_AttrUL li").Each(func(index int, element *goquery.Selection) {
		as := strings.Split(element.Text(), ":")
		if len(as) < 2 {
			as = strings.Split(element.Text(), "：")
		}

		b := ""

		if len(as) >= 2 && !utf8.ValidString(as[1]) {
			as[1] = as[1]
			b = as[1]
		}

		attribute = append(attribute, as[0]+":"+b)
	})

	if len(attribute) == 0 {
		p.Find("#attributes .attributes-list li").Each(func(index int, element *goquery.Selection) {
			attribute = append(attribute, element.Text())
		})
	}

	return strings.Join(attribute, "##")
}

//获取分类ID
func GetCategoryId(h string) string {
	regTao, _ := regexp.Compile(`category=item%5f(\d+)`)
	if res := regTao.FindAllStringSubmatch(h, -1); len(res) > 0 {
		return strings.TrimSpace(res[0][1])
	}
	return ""
}

//获取淘宝客的分类和gid
func GetTaoCategoryId(h string) (catid string, gid string) {
	gidReg, _ := regexp.Compile(`"itemId":"(\d+)"`)
	if res := gidReg.FindAllStringSubmatch(h, -1); len(res) > 0 {
		gid = strings.TrimSpace(res[0][1])
	}

	catidReg, _ := regexp.Compile(`"catid":"(\d+)"`)
	if res := catidReg.FindAllStringSubmatch(h, -1); len(res) > 0 {
		catid = strings.TrimSpace(res[0][1])
	}

	return
}

//设置cookie
func SetGrabCookie(ck string) {

}

//
func SetUserAgent(ua string) {

}

func SetTransport(proxyurl string) {

}

//获取店铺信息
func GetShopId(p *goquery.Document) string {
	str, _ := p.Find(`meta[name="microscope-data"]`).Attr("content")

	//str := metas.Eq(aindex).Attr("content")

	reg, _ := regexp.Compile(`shopId=(\d+);`)
	res := reg.FindStringSubmatch(str)
	if len(res) >= 1 {
		return strings.TrimSpace(res[1])
	}
	return ""
}

//获取店铺名称
func GetShopName(p *goquery.Document) string {
	name := p.Find(".tb-shop-name").Text()
	if name == "" {
		name = p.Find(".slogo-shopname").Text()
	}
	return strings.TrimSpace(name)
}

//获取店铺地址
func GetShopUrl(p *goquery.Document) string {
	href, _ := p.Find(".tb-seller-name").Attr("href")
	if href == "" {
		href, _ = p.Find(".slogo-shopname").Attr("href")
	}
	return strings.TrimSpace("https:" + href)
}

//获取店铺掌柜
func GetShopBoss(p *goquery.Document) string {
	name := p.Find(".tb-seller-name").Text()
	if name != "" {
		return strings.TrimSpace(name)
	}

	reg, _ := regexp.Compile(`"sellerNickName":"([\w%%]+)"`)
	a, _ := p.Html()
	res := reg.FindStringSubmatch(a)

	if len(res) >= 1 {
		b, _ := url.QueryUnescape(res[1])
		return strings.TrimSpace(b)
	}
	return ""
}

// 获取店铺id通过店铺地址
func GetShopIdByShop(p *goquery.Document) string {
	v, e := p.Find("#LineZing").Attr("shopid")
	if e {
		return strings.TrimSpace(v)
	}

	return ""
}

//获取店铺名称通过店铺地址
func GetShopNameByShop(p *goquery.Document) string {
	return strings.TrimSpace(p.Find(".shop-name span").Text())
}

//获取店铺掌柜通过店铺地址
func GetShopBossByShop(h string) (boss string) {
	//"user_nick": "%E8%98%91%E8%8F%87%E8%A1%97%E5%86%AC%E8%A3%85%E5%A5%B3%E8%A3%85"
	reg, _ := regexp.Compile(`"user_nick"\:\s+"([%%\w]+)"`)
	res := reg.FindStringSubmatch(h)
	if len(res) >= 1 {
		boss, _ = url.QueryUnescape(res[1])
		boss = strings.TrimSpace(boss)
		return
	}
	return ""
}

func In_Array(list []string, k string) bool {
	for _, v := range list {
		if v == k {
			return true
		}
	}
	return false
}

func In_Array_Array(list []string, k []string) []string {
	var plist []string = make([]string, 0, 100)
	for _, v := range list {
		for _, vv := range k {
			if v == vv {
				plist = append(plist, vv)
			}
		}
	}
	return plist
}

package grab

import (
	"crypto/tls"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"unicode/utf8"

	//"github.com/opesun/goquery"
	//"github.com/opesun/goquery/exp/html"
	"github.com/PuerkitoBio/goquery"
	"github.com/qiniu/iconv"
)

var (
	cd        iconv.Iconv
	cookie    string
	useragent string = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/42.0.2311.152 Safari/537.36"
	client    *http.Client
)

func init() {
	cd, _ = iconv.Open("utf-8", "gbk")
	client = &http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}}
}

//解析页面到node
func ParseNode(h string) (*goquery.Document, error) {
	return goquery.NewDocumentFromReader(strings.NewReader(h))
}

//抓取淘宝商品页面
func GrabTaoHTML(url string) string {
	r, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return ""
	}
	r.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	//r.Header.Set("Accept-Encoding", "gzip, deflate, sdch")
	r.Header.Set("Accept-Language", "zh-CN,zh;q=0.8")
	r.Header.Set("Cache-Control", "no-cache")
	r.Header.Set("Connection", "keep-alive")
	//r.Header.Set("Cookie", cookie)
	r.Header.Set("Host", "www.taobao.com")
	r.Header.Set("Pragma", "no-cache")
	r.Header.Set("User-Agent", useragent)

	resp, err := client.Do(r)
	if err != nil {
		if resp == nil {
			log.Println("================")
			log.Println(url)
			log.Println("================")
			log.Println(err)
			return ""
		}
	}

	defer resp.Body.Close()

	if resp.StatusCode == 302 || resp.StatusCode == 301 {
		location := resp.Header.Get("Location")
		return GrabTaoHTML(location)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return ""
	}

	return string(body)
}

//获取淘宝标题
func GetTitle(p *goquery.Document) string {
	return cd.ConvString(p.Find("title").Text())
}

//获取淘宝属性信息
func GetAttrbuites(p *goquery.Document) string {
	attribute := make([]string, 0, 20)
	p.Find("#J_AttrUL li").Each(func(index int, element *goquery.Selection) {
		as := strings.Split(element.Text(), ":")

		if len(as) < 2 {
			as = strings.Split(cd.ConvString(element.Text()), "：")
		}
		if !utf8.ValidString(as[0]) {
			as[0] = cd.ConvString(as[0])
		}

		b := ""

		if len(as) >= 2 && !utf8.ValidString(as[1]) {
			as[1] = cd.ConvString(as[1])
			b = as[1]
		}

		attribute = append(attribute, as[0]+":"+b)
	})

	if len(attribute) == 0 {
		p.Find("#attributes .attributes-list li").Each(func(index int, element *goquery.Selection) {
			attribute = append(attribute, cd.ConvString(element.Text()))
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
	cookie = ck
}

//
func SetUserAgent(ua string) {
	useragent = ua
}

func SetTransport(proxyurl string) {
	proxy, err := url.Parse(proxyurl)
	if err != nil {
		return
	}

	if v, ok := client.Transport.(*http.Transport); ok {
		v.Proxy = http.ProxyURL(proxy)
	}
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
	return strings.TrimSpace(cd.ConvString(name))
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
		return strings.TrimSpace(cd.ConvString(name))
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
	return strings.TrimSpace(cd.ConvString(p.Find(".shop-name span").Text()))
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

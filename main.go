package main

import (
	"fmt"

	"github.com/goweb/gopro/lib/grab"
)

func main() {
	//	os.Setenv("SURF_DEBUG_HEADERS", "111")
	//	cd, _ := iconv.Open("utf-8", "gbk")
	//	d := surfer.NewDownload(2, time.Second, "")
	//	for {
	//		d.SetUserAgent(agent.Firefox())
	//		r, _ := d.Get("https://item.taobao.com/item.htm?id=43873054934", nil, nil)
	//		dd, _ := goquery.NewDocumentFromResponse(r)
	//		h, _ := dd.Find("title").Html()
	//		fmt.Println(cd.ConvString(h))

	//		h = grab.GrabTaoHTML("https://item.taobao.com/item.htm?id=43873054934")
	//		n, _ := grab.ParseNode(h)
	//		fmt.Println(grab.GetTitle(n))
	//	}
	for {
		h := grab.GrabTaoHTML("https://item.taobao.com/item.htm?id=43873054934")
		n, _ := grab.ParseNode(h)
		fmt.Println(grab.GetTitle(n))
	}
}

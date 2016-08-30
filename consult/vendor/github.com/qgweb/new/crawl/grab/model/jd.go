package model

import "github.com/qgweb/new/lib/grab"

type JDCrawl struct {
}

func (this JDCrawl) Grab(gid string) map[string]interface{} {
	var info = make(map[string]interface{})
	url := "http://item.jd.com/" + gid + ".html"
	h := grab.GrabJDHTML(url)
	if h == "" {
		return nil
	}
	p, _ := grab.ParseNode(h)
	//标题
	info["title"] = grab.GetJDTitle(p)
	//分类
	cat := grab.GetJDCategory(p)
	info["cat_id"] = cat[1]
	info["cat_name"] = cat[0]
	//品牌
	info["brand"] = grab.GetJDBrand(p)
	//属性
	info["attributes"] = grab.GetJDAttributes(p)
	//id
	info["gid"] = gid
	return info
}

package model

import (
	"github.com/ngaut/log"
	"github.com/qgweb/new/lib/grab"
	"strings"
	"time"
)

var (
	sexmap    map[string]int = map[string]int{"中性": 0, "男": 1, "女": 2}
	peoplemap map[string]int = map[string]int{"青年": 0, "孕妇": 1, "中老年": 2, "儿童": 3, "青少年": 4, "婴儿": 5}
)

type TaobaoCrawl struct {
}

func (this TaobaoCrawl) Grab(gid string) map[string]interface{} {
LABEL:
	url := "https://item.taobao.com/item.htm?id=" + gid
	h := grab.GrabTaoHTML(url)

	if h == "" {
		return nil
	}

	p, _ := grab.ParseNode(h)

	//标签名称
	title := grab.GetTitle(p)

	if title == "淘宝网 - 淘！我喜欢" || strings.Contains(title, "出错啦！") {
		//log.Println("商品不存在,id为:", gid)
		return nil
	}

	if strings.Contains(title, "访问受限") {
		log.Error("访问受限,id为", gid)
		time.Sleep(time.Minute * 2)
		goto LABEL
		return nil
	}

	//标签id
	cateId := grab.GetCategoryId(h)

	//标签信息

	//特性
	features := make(map[string]int)

	//属性
	attrbuites := grab.GetAttrbuites(p)

	//性别
	sex := 0
	for k, v := range sexmap {
		if strings.Contains(title, k) {
			sex = v
			break
		}
	}
	//人群
	people := 0
	for k, v := range peoplemap {
		if strings.Contains(title, k) {
			people = v
			break
		}
	}
	//品牌
	brand := grab.GetBrand(attrbuites)

	// 店铺信息
	shopId := grab.GetShopId(p)
	shopName := grab.GetShopName(p)
	shopUrl := grab.GetShopUrl(p)
	shopBoss := grab.GetShopBoss(p)

	return map[string]interface{}{
		"shop_id":   shopId,
		"shop_name": shopName,
		"shop_url":  shopUrl,
		"shop_boss": shopBoss,
		"gid":       gid,
		//"tagname":    cateInfo["name"],
		"tagid":      cateId,
		"features":   features,
		"attrbuites": attrbuites,
		"sex":        sex,
		"people":     people,
		"brand":      brand,
	}
}

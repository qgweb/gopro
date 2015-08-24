//抓取淘宝数据
package main

import (
	"gopro/lib/grab"
	"log"
	"strings"
	"time"

	gs "goclass/grab"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var (
	sexmap    map[string]int = map[string]int{"中性": 0, "男": 1, "女": 2}
	peoplemap map[string]int = map[string]int{"青年": 0, "孕妇": 1, "中老年": 2, "儿童": 3, "青少年": 4, "婴儿": 5}
)

//获取淘宝分类
func GetTaoCat(cid string) (map[string]interface{}, error) {
	sess := GetSession()
	defer sess.Close()
	info := make(map[string]interface{})
	err := sess.DB(modb).C("taocat").Find(bson.M{"cid": cid}).One(&info)
	if err != nil {
		return nil, err
	}
	return info, nil
}

//添加商品
func AddGoodsInfo(gid string) (info map[string]string) {
	url := "http://item.taobao.com/item.htm?id=" + gid
	grab.SetGrabCookie(cookie)
	h := grab.GrabTaoHTML(url)
	p, _ := grab.ParseNode(h)

	//标签名称
	title := grab.GetTitle(p)
	if title == "淘宝网 - 淘！我喜欢" || strings.Contains(title, "出错啦！") {
		//log.Println("商品不存在,id为:", gid)
		return nil
	}

	if strings.Contains(title, "访问受限") {
		log.Println("访问受限,id为", gid)
		time.Sleep(time.Minute * 2)
		return nil
	}

	//标签id
	cateId := grab.GetCategoryId(h)

	cateInfo, err := GetTaoCat(cateId)
	if err == mgo.ErrNotFound {
		//log.Println("分类ID:", gid, "-", cateId, "-", err)
		return nil
	}
	//特性
	features := make(map[string]int)
	if v, ok := cateInfo["features"]; ok {
		for a, b := range v.(map[string]interface{}) {
			features[a] = b.(int)
		}
	}

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

	// 店铺信息
	shopId := grab.GetShopId(p)
	shopName := grab.GetShopName(p)
	shopUrl := grab.GetShopUrl(p)
	shopBoss := grab.GetShopBoss(p)

	//浏览数
	count := 0
	go func() {
		sess := GetSession()
		defer func() {
			sess.Close()
		}()

		sess.DB(modb).C("goods").Upsert(bson.M{"gid": gid}, bson.M{"$set": bson.M{
			"tagname": cateInfo["name"], "tagid": cateId, "features": features,
			"attrbuites": attrbuites, "sex": sex, "people": people,
			"shop_id": shopId, "shop_name": shopName, "shop_url": shopUrl,
			"shop_box": shopBoss, "count": count}})
	}()

	return map[string]string{
		"cid":       cateId,
		"shop_id":   shopId,
		"shop_name": shopName,
		"shop_url":  shopUrl,
		"shop_boss": shopBoss,
	}
}

//添加用户id对应分类id
func AddUidCids(param map[string]string) {
	//ad string, cids string, cookie string, ua string
	sess := GetSession()
	defer func() {
		sess.Close()
	}()

	tableName := prefix + "ad_tags"
	if param["cookie"] != "" {
		tableName = prefix + "cookie_tags"
	}

	//无cookie情况
	if param["cookie"] == "" {
		sess.DB(modb).C(tableName).Upsert(bson.M{"ad": param["ad"], "ua": param["ua"]},
			bson.M{"$set": bson.M{"cids": param["cids"]}})

		//按小时存储
		t := tableName + "_clock"
		c := "cids." + param["clock"]
		d := param["date"]
		sess.DB(modb).C(t).Upsert(bson.M{"ad": param["ad"], "ua": param["ua"], "date": d},
			bson.M{"$set": bson.M{c: param["cids"]}})

		//存储对应的店铺信息
		t = tableName + "_shop"
		for _, v := range strings.Split(param["shops"], ",") {
			sess.DB(modb).C(t).Upsert(bson.M{"ad": param["ad"], "ua": param["ua"], "date": d},
				bson.M{"$addToSet": bson.M{"shop": bson.M{"id": v}}})
		}

	} else {
		sess.DB(modb).C(tableName).Upsert(bson.M{"cookie": param["cookie"]},
			bson.M{"$set": bson.M{"cids": param["cids"], "ad": param["ad"]}})

		//统计标签频率
		tagFrequencyRecord(param["cookie"], param["cids"])
	}
}

// 标签频率统计记录
func tagFrequencyRecord(cookie string, cids string) {
	sess := GetSession()
	defer sess.Close()

	//分割标签
	tagAry := strings.Split(cids, ",")
	tagsMap := make(map[string]int)

	for _, v := range tagAry {
		if _, ok := tagsMap[v]; ok {
			tagsMap[v] = tagsMap[v] + 1
		} else {
			tagsMap[v] = 1
		}
	}

	//排序
	s := gs.NewMapSorter(tagsMap)
	s.Sort()

	bms := make([]bson.M, 0, 20)
	for _, v := range s {
		bms = append(bms, bson.M{"tagid": v.Key, "tagcount": v.Val})
	}

	//插入mongo
	tableName := prefix + "cookie_tags_put"
	sess.DB(modb).C(tableName).Upsert(bson.M{"cookie": cookie},
		bson.M{"cookie": cookie, "cids": bms, "date": time.Now().Format("2006-01-02 15:04:05")})
}

//添加店铺
func AddShopInfo(param map[string]string) {
	sess := GetSession()
	defer sess.Close()

	sess.DB(modb).C("taoshop").Upsert(bson.M{"shop_id": param["shop_id"]}, bson.M{
		"$set": bson.M{
			"shop_name": param["shop_name"],
			"shop_url":  param["shop_url"],
			"shop_boss": param["shop_boss"],
		},
	})
}

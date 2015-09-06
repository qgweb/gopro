package main

import (
	"errors"
	"github.com/qgweb/gopro/lib/grab"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"regexp"
	"strings"
)

//店铺结构
type ShopInfo struct {
	ShopName string `json:"shop_name"`
	ShopId   string `json:"shop_id"`
	ShopUrl  string `json:"shop_url"`
	ShopBoss string `json:"shop_boss"`
}

//淘宝店铺rpc接口
type TaoShop struct {
}

//抓取店铺信息
func (this TaoShop) grabShopInfo(shop_url string) (si ShopInfo, err error) {
	var getTaoId = func(url string) string {
		reg, _ := regexp.Compile(`shop(\d+)\.`)
		res := reg.FindStringSubmatch(url)
		if len(res) >= 1 {
			return res[1]
		}

		return ""
	}

	var getShopId = func(url string, f func() string) string {
		//判断是否是天猫
		if strings.Contains(url, "tmall") {
			return f()
		}

		if getTaoId(url) == "" {
			return f()
		}

		return getTaoId(url)
	}

	h := grab.GrabTaoHTML(shop_url)
	if h == "" {
		return si, errors.New("抓取数据为空")
	}

	node, err := grab.ParseNode(h)
	if err != nil {
		return si, err
	}

	si.ShopId = getShopId(shop_url, func() string {
		return grab.GetShopIdByShop(node)
	})
	si.ShopBoss = grab.GetShopBossByShop(h)
	si.ShopName = grab.GetShopNameByShop(node)
	si.ShopUrl = shop_url
	return si, nil
}

//获取店铺信息
func (this TaoShop) GetShopInfo(param string) []byte {
	//param : 店铺地址,旺旺id, 店铺中文名
	var (
		modb = IniFile.Section("mongo-xu_precise").Key("db").String()
		info map[string]interface{}
	)
	sess := GetSession()
	defer sess.Close()

	//db.taoshop.find({ "$or" : [{shop_name : "尔岚服饰旗舰店"}, {"shop_url": "https://erlan.tmall.com"}, {"shop_boss": "尔岚服饰旗舰店"} ]})
	err := sess.DB(modb).C("taoshop").Find(bson.M{"$or": []bson.M{
		bson.M{"shop_name": param},
		bson.M{"shop_url": param},
		bson.M{"shop_boss": param},
	}}).Select(bson.M{"_id": 0}).One(&info)

	if err == mgo.ErrNotFound {
		//抓取店铺信息
		if strings.Contains(param, "http") {
			si, err := this.grabShopInfo(param)
			if err != nil {
				return jsonReturn(nil, errors.New("抓取店铺信息失败"))
			}
			// 添加店铺
			this.addShop(si)
			return jsonReturn(si, nil)
		}
		return jsonReturn(nil, errors.New("数据不存在"))
	}

	return jsonReturn(info, nil)
}

// 添加商铺
func (this TaoShop) addShop(si ShopInfo) {
	var (
		modb = IniFile.Section("mongo-xu_precise").Key("db").String()
	)

	sess := GetSession()
	defer sess.Close()

	sess.DB(modb).C("taoshop").Upsert(bson.M{"shop_id": si.ShopId}, bson.M{
		"$set": bson.M{
			"shop_name": si.ShopName,
			"shop_url":  si.ShopUrl,
			"shop_boss": si.ShopBoss,
		}})
}

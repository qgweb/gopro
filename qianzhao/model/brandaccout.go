package model

import (
	"log"

	"github.com/qgweb/gopro/lib/convert"
)

const (
	BROAD_TABLE_NAME = "211su_broadband"
)

//宽带账户|属地|
//校园名称|校园组别|上行带宽|下
//行带宽
type BrandAccount struct {
	Id            string `json:"id"`             //编号
	Area          string `json:"area"`           //属地（带争议）
	Account       string `json:"account"`        //宽带账户
	SchoolName    string `json:"school_name"`    //校园名称
	SchoolGroup   string `json:"school_group"`   //校园组别
	UpBroadband   string `json:"broadband_up"`   //上行带宽
	DownBroadband string `json:"broadband_down"` //下行带宽
}

// 添加白名单
func (this *BrandAccount) AddBroadBand(ba BrandAccount) bool {
	myorm.BSQL().Insert(BROAD_TABLE_NAME).Values("area", "account", "school_name",
		"school_group", "broadband_up", "broadband_down")
	n, err := myorm.Insert(ba.Area, ba.Account, ba.SchoolName, ba.SchoolGroup, ba.UpBroadband, ba.DownBroadband)
	if err != nil {
		log.Println("[brandaccount AddBroadBand] 添加白名单失败 ", err)
		return false
	}

	if n > 0 {
		return true
	}

	return false
}

// 判断账户是否存在
func (this *BrandAccount) AccountExist(ba BrandAccount) bool {
	myorm.BSQL().Select("id").From(BROAD_TABLE_NAME).Where("account=?")
	list, err := myorm.Query(ba.Account)
	if err != nil {
		log.Println("[brandaccount AccountExist] 查询失败，", err)
		return false
	}

	if len(list) > 0 && convert.ToInt(list[0]["id"]) > 0 {
		return true
	}

	return false
}

// 更新账户
func (this *BrandAccount) EditAccount(ba BrandAccount) bool {
	myorm.BSQL().Update(BROAD_TABLE_NAME).Set("area", "school_name", "school_group",
		"broadband_up", "broadband_down").Where("account=?")
	n, err := myorm.Update(ba.Area, ba.SchoolName, ba.SchoolGroup, ba.UpBroadband, ba.DownBroadband)
	if err != nil {
		log.Println("[brandaccount EditAccount] 更新失败，", err)
		return false
	}
	if n > 0 {
		return true
	}
	return false
}

package model

import (
	"log"

	"github.com/qgweb/gopro/lib/convert"
)

const (
	BROAD_TABLE_NAME = "221su_broadband"
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
	UpBroadband   int    `json:"broadband_up"`   //上行带宽
	DownBroadband int    `json:"broadband_down"` //下行带宽
	TotalTime     int    `json:"total_time"`     //总共体验时长
	UsedTime      int    `json:"used_time"`      //已体验时长
	TryCount      int    `json:"try_count"`      // 已尝试次数
}

// 添加白名单
func (this *BrandAccount) AddBroadBand(ba BrandAccount) bool {
	myorm.BSQL().Insert(BROAD_TABLE_NAME).Values("area", "account", "school_name",
		"school_group", "broadband_up", "broadband_down", "total_time", "used_time", "try_count")
	n, err := myorm.Insert(ba.Area, ba.Account, ba.SchoolName, ba.SchoolGroup,
		ba.UpBroadband, ba.DownBroadband, ba.TotalTime, ba.UsedTime, ba.TryCount)
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
	log.Println(list, err, ba.Account, myorm.LastSql())
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
		"broadband_up", "broadband_down", "used_time", "total_time", "used_time", "try_count").Where("account=?")
	n, err := myorm.Update(ba.Area, ba.SchoolName, ba.SchoolGroup,
		ba.UpBroadband, ba.DownBroadband, ba.UsedTime, ba.TotalTime,
		ba.UsedTime, ba.TryCount, ba.Account)
	if err != nil {
		log.Println("[brandaccount EditAccount] 更新失败，", err)
		return false
	}
	if n > 0 {
		return true
	}
	return false
}

// 获取用户信息
func (this *BrandAccount) GetAccountInfo(account string) (BrandAccount, error) {
	myorm.BSQL().Select("*").From(BROAD_TABLE_NAME).Where("account=?")
	list, err := myorm.Query(account)
	if err != nil {
		log.Println("[brandaccount GetAccountInfo] 查询失败，", err)
		return BrandAccount{}, err
	}

	if len(list) == 0 {
		return BrandAccount{}, nil
	}

	ba := BrandAccount{}
	ba.Id = list[0]["id"]
	ba.Account = list[0]["account"]
	ba.Area = list[0]["area"]
	ba.SchoolName = list[0]["school_name"]
	ba.SchoolGroup = list[0]["school_group"]
	ba.UpBroadband = convert.ToInt(list[0]["broadband_up"])
	ba.DownBroadband = convert.ToInt(list[0]["broadband_down"])
	ba.TotalTime = convert.ToInt(list[0]["total_time"])
	ba.UsedTime = convert.ToInt(list[0]["used_time"])
	ba.TryCount = convert.ToInt(list[0]["try_count"])

	return ba, nil
}

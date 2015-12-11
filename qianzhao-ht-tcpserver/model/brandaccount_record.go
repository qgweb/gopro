package model

import (
	"time"

	"github.com/ngaut/log"
	"github.com/qgweb/gopro/lib/convert"
)

const (
	BROADRECORD_TABLE_NAME = "221su_broadband_dz_record"
	SPEED_UP_TIME          = 3600
)

type BrandAccountRecord struct {
	Id        int64  `json:"id"`         //编号
	Account   string `json:"account"`    //宽带账户
	BeginTime int64  `json:"begin_time"` //开始时间
	EndTime   int64  `json:"end_time"`   //结束时间
	Date      int64  `json:"date"`       //日期
}

func (this *BrandAccountRecord) AddRecord(info BrandAccountRecord) bool {
	myorm.BSQL().Insert(BROADRECORD_TABLE_NAME).Values("account", "begin_time", "end_time", "date")
	n, err := myorm.Insert(info.Account, info.BeginTime, info.EndTime, info.Date)
	if err != nil {
		log.Error("[model BrandAccountRecord AddRecord] 插入记录失败 ", err)
		return false
	}

	if n > 0 {
		return true
	}
	return false
}

// 获取账号剩余时间
func (this *BrandAccountRecord) GetAccountCanUserTime(account string) int {
	date, _ := time.Parse("2006-01-02", time.Now().Format("2006-01-02"))
	myorm.BSQL().Select("sum(end_time - begin_time) as time").From(BROADRECORD_TABLE_NAME).
		Where("date=? and account=?")
	list, err := myorm.Query(date.Unix(), account)
	if err != nil {
		log.Error("[model BrandAccountRecord GetAccountCanUserTime] 查询失败,", err)
		return 0
	}

	if len(list) == 0 {
		return SPEED_UP_TIME
	}
	return SPEED_UP_TIME - convert.ToInt(list[0]["time"])
}

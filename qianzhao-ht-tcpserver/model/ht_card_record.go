package model

import (
	"time"

	"github.com/ngaut/log"
	"github.com/qgweb/gopro/lib/convert"
)

const (
	HT_CARD_RECORD_TABLE_NAME = "221su_ht_card_record"
	HT_SPEED_UP_TIME          = 60 //免费版提速时间
)

type HtCardRecord struct {
	Id        int64 `json:"id"`         //编号
	HtId      int64 `json:"ht_id"`      //卡申请表号
	BeginTime int64 `json:"begin_time"` //开始时间
	EndTime   int64 `json:"end_time"`   //结束时间
	Date      int64 `json:"date"`       //日期
}

func (this *HtCardRecord) AddRecord(info HtCardRecord) bool {
	myorm.BSQL().Insert(HT_CARD_RECORD_TABLE_NAME).Values("ht_id", "begin_time", "end_time", "date")
	n, err := myorm.Insert(info.HtId, info.BeginTime, info.EndTime, info.Date)
	if err != nil {
		log.Error("[model HtCardRecord AddRecord] 插入记录失败 ", err)
		return false
	}

	if n > 0 {
		return true
	}
	return false
}

// 获取账号剩余时间
func (this *HtCardRecord) GetAccountCanUserTime(HtId int, totalTime int) int {
	date, _ := time.Parse("2006-01-02", time.Now().Format("2006-01-02"))
	myorm.BSQL().Select("sum(end_time - begin_time) as time").From(HT_CARD_RECORD_TABLE_NAME).
		Where("date=? and ht_id=? and end_time <> 0")
	list, err := myorm.Query(date.Unix(), HtId)
	if err != nil {
		log.Error("[model HtCardRecord GetAccountCanUserTime] 查询失败,", err)
		return 0
	}

	if len(list) == 0 {
		return totalTime
	}
	return totalTime - convert.ToInt(list[0]["time"])
}

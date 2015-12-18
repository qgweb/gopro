package model

import (
	"github.com/ngaut/log"
	"github.com/qgweb/gopro/lib/convert"
)

const (
	HT_CARD_TABLE_NAME = "221su_ht_card"
)

// 海淘申请卡
type HTCard struct {
	Id        int    `json:id`
	Phone     string `json:phone`      //手机号码
	CardNum   string `json:card_num`   //卡号
	CardPwd   string `json:card_pwd`   //密码
	CardToken string `json:card_token` //令牌
	CardType  int    `json:type`       //卡类型
	TotalTime int    `json:total_time` //可使用时间
	Status    int    `json:status`     //卡状态
	Date      int    `json:date`       //日期
	Remark    string `json:remark`     //备注
}

// 添加记录
func (this *HTCard) AddReocrd(info HTCard) int {
	myorm.BSQL().Insert(HT_CARD_TABLE_NAME).Values("phone", "card_num",
		"card_pwd", "card_token", "type", "total_time", "status", "date",
		"remark")
	n, err := myorm.Insert(info.Phone, info.CardNum, info.CardPwd, info.CardToken,
		info.CardType, info.TotalTime, info.Status, info.Date, info.Remark)
	if err != nil {
		log.Error(err)
		return 0
	}

	if n > 0 {
		return int(n)
	}

	return 0
}

func (this *HTCard) getCard(info map[string]string) (ht HTCard) {
	if v, ok := info["id"]; ok {
		ht.Id = convert.ToInt(v)
	}
	if v, ok := info["phone"]; ok {
		ht.Phone = convert.ToString(v)
	}
	if v, ok := info["card_num"]; ok {
		ht.CardNum = convert.ToString(v)
	}
	if v, ok := info["card_pwd"]; ok {
		ht.CardPwd = convert.ToString(v)
	}
	if v, ok := info["card_token"]; ok {
		ht.CardToken = convert.ToString(v)
	}
	if v, ok := info["type"]; ok {
		ht.CardType = convert.ToInt(v)
	}
	if v, ok := info["total_time"]; ok {
		ht.TotalTime = convert.ToInt(v)
	}
	if v, ok := info["status"]; ok {
		ht.Status = convert.ToInt(v)
	}
	if v, ok := info["date"]; ok {
		ht.Date = convert.ToInt(v)
	}
	if v, ok := info["remark"]; ok {
		ht.Remark = convert.ToString(v)
	}
	return
}

// 获取信息
func (this *HTCard) GetInfo(id int) (ht HTCard) {
	myorm.BSQL().Select("*").From(HT_CARD_TABLE_NAME).Where("id=?")
	list, err := myorm.Query(id)
	if err != nil {
		log.Error(err)
		return ht
	}
	if len(list) > 0 {
		return this.getCard(list[0])
	}
	return ht
}

// 获取信息
func (this *HTCard) GetInfoByPhone(phone string, date string, ctype int, status int) (ht HTCard) {
	myorm.BSQL().Select("*").From(HT_CARD_TABLE_NAME).Where("phone=? and date=? and type = ? and status=?")
	list, err := myorm.Query(phone, date, ctype, status)
	if err != nil {
		log.Error(err)
		return ht
	}
	if len(list) > 0 {
		return this.getCard(list[0])
	}
	return ht
}

// 获取信息
func (this *HTCard) GetInfoByCard(phone string, cardno string, status int) (ht HTCard) {
	myorm.BSQL().Select("*").From(HT_CARD_TABLE_NAME).Where("phone=?  and card_num = ? and status=?").
		Order("date desc").Limit(1)
	list, err := myorm.Query(phone, cardno, status)
	if err != nil {
		log.Error(err)
		return ht
	}
	if len(list) > 0 {
		return this.getCard(list[0])
	}
	return ht
}

// 获取最近的卡信息
func (this *HTCard) GetMoneyLastCard(phone string) (ht HTCard) {
	myorm.BSQL().Select("*").From(HT_CARD_TABLE_NAME).Where("phone=? and type=? and status=?").Order("date desc").Limit(1)
	list, err := myorm.Query(phone, 1, 1)
	if err != nil {
		log.Error(err)
		return ht
	}
	if len(list) > 0 {
		return this.getCard(list[0])
	}
	return ht
}

// 更新卡信息
func (this *HTCard) UpdateCard(info HTCard) bool {
	myorm.BSQL().Update(HT_CARD_TABLE_NAME).Set("phone", "card_num",
		"card_pwd", "card_token", "type", "total_time", "status", "date",
		"remark")
	n, err := myorm.Update(info.Phone, info.CardNum, info.CardPwd, info.CardToken,
		info.CardType, info.TotalTime, info.Status, info.Date, info.Remark)
	if err != nil {
		log.Error(err)
		return false
	}

	if n > 0 {
		return true
	}

	return false
}

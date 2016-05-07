package model

import (
	"log"
)

const (
	BROADRECORD_TABLE_NAME = "221su_broadband_record"
)

// CREATE TABLE `211su_broadband_record` (
//   `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
//   `account` varchar(50) NOT NULL DEFAULT '' COMMENT '宽带账号',
//   `begin_time` int(11) NOT NULL DEFAULT '0' COMMENT '开始时间',
//   `end_time` int(11) NOT NULL DEFAULT '0' COMMENT '结束时间',
//   `time` int(11) NOT NULL DEFAULT '0' COMMENT '创建时间',
//   PRIMARY KEY (`id`)
// ) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='宽带使用记录表';

type BrandAccountRecord struct {
	Id        int64  `json:"id"`         //编号
	Account   string `json:"account"`    //宽带账户
	BeginTime int64  `json:"begin_time"` //开始时间
	EndTime   int64  `json:"end_time"`   //结束时间
	Time      int64  `json:"time"`       //创建时间
}

func (this *BrandAccountRecord) AddRecord(info BrandAccountRecord) bool {
	sql := myorm.BSQL().Insert(BROADRECORD_TABLE_NAME).Values("account", "begin_time", "end_time", "time").GetSQL()
	n, err := myorm.Insert(sql, info.Account, info.BeginTime, info.EndTime, info.Time)
	if err != nil {
		log.Println("[model BrandAccountRecord AddRecord] 插入记录失败 ", err)
		return false
	}

	if n > 0 {
		return true
	}
	return false
}

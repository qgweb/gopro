package model

import (
	"github.com/qgweb/new/lib/convert"
	"github.com/qgweb/new/lib/timestamp"
)

type AwardCount struct {
	ID     int
	Time   int
	Five   int
	Ten    int
	Twenty int
}

const (
	AWARDCount_TABLE_NAME = "221su_award_count"
)

func (this *AwardCount) To(v map[string]string) (ar AwardCount) {
	if vv, ok := v["id"]; ok {
		ar.ID = convert.ToInt(vv)
	}
	if vv, ok := v["time"]; ok {
		ar.Time = convert.ToInt(vv)
	}
	if vv, ok := v["five"]; ok {
		ar.Five = convert.ToInt(vv)
	}
	if vv, ok := v["ten"]; ok {
		ar.Ten = convert.ToInt(vv)
	}
	if vv, ok := v["twenty"]; ok {
		ar.Twenty = convert.ToInt(vv)
	}
	return
}

func (this *AwardCount) Get() (ar AwardCount, err error) {
	sql := myorm.BSQL().Select("*").From(AWARDCount_TABLE_NAME).Where("time=?").GetSQL()
	info, err := myorm.Get(sql, timestamp.GetDayTimestamp(0))
	if err != nil {
		return ar, err
	}
	return this.To(info), err
}

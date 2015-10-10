package model

import (
	"github.com/ngaut/log"
	"github.com/qgweb/gopro/lib/convert"
)

const (
	TABLE_NAME_DAYACTIVITY = "221su_day_activity"
)

type DayActivity struct {
	Id      int    `json:"id"`
	Version string `json:"version"`
	Count   int    `json:"count"`
	Type    int    `json:"type"`
	Date    int64  `json:"date"`
}

func (this *DayActivity) AddRecord(dl *DayActivity) bool {
	//查找是否存在
	myorm.BSQL().Select("id", "count").From(TABLE_NAME_DAYACTIVITY).Where("version=? and type=? and date=?")
	list, err := myorm.Query(dl.Version, dl.Type, dl.Date)
	if err != nil {
		log.Error(err)
		return false
	}

	if len(list) > 0 {
		myorm.BSQL().Update(TABLE_NAME_DAYACTIVITY).Set("count").Where("id=?")
		n, err := myorm.Update(convert.ToInt(list[0]["count"])+1, list[0]["id"])
		if err != nil {
			log.Error(err)
			return false
		}
		if n > 0 {
			return true
		}
		return false

	} else {
		myorm.BSQL().Insert(TABLE_NAME_DAYACTIVITY).Values("version", "count", "type",
			"date")

		n, err := myorm.Insert(dl.Version, 1, dl.Type, dl.Date)
		if err != nil {
			log.Error(err)
			return false
		}
		if n > 0 {
			return true
		}
		return false
	}

}

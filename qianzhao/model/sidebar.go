package model

import (
	"github.com/ngaut/log"
	"github.com/qgweb/gopro/lib/convert"
)

const (
	TABLE_NAME_SIDEBAR = "221su_sidebar"
)

type SideBar struct {
	Id       int    `json:"id"`
	Favorite int    `json:"favorite"`
	Email    int    `json:"email"`
	Type     int    `json:"type"`
	Yixin    int    `json:"yixin"`
	Date     int64  `json:"date"`
	Version  string `json:"version"`
}

func (this *SideBar) AddRecord(dl *SideBar) bool {
	//查找是否存在
	myorm.BSQL().Select("*").From(TABLE_NAME_SIDEBAR).Where("date=? and type=? and version=?")
	list, err := myorm.Query(dl.Date, dl.Type)
	if err != nil {
		log.Error(err)
		return false
	}

	if len(list) > 0 {
		myorm.BSQL().Update(TABLE_NAME_SIDEBAR).Set("favorite", "email", "yixin").Where("id=?")
		n, err := myorm.Update(convert.ToInt(list[0]["favorite"])+dl.Favorite,
			convert.ToInt(list[0]["email"])+dl.Email,
			convert.ToInt(list[0]["yixin"])+dl.Yixin,
			list[0]["id"])
		if err != nil {
			log.Error(err)
			return false
		}
		if n > 0 {
			return true
		}
		return false

	} else {
		myorm.BSQL().Insert(TABLE_NAME_SIDEBAR).Values("favorite", "email",
			"yixin", "type", "date", "version")

		n, err := myorm.Insert(1, 1, 1, dl.Type, dl.Date, dl.Version)
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

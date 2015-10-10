package model

import (
	"github.com/ngaut/log"
)

const (
	TABLE_NAME_DOWNLOAD = "221su_download"
)

type Download struct {
	Id      int    `json:"id"`
	Version string `json:"version"`
	Channel int    `json:"channel"`
	Type    int    `json:"type"`
	Account string `json:"account"`
	City    string `json:"city"`
	Date    int64  `json:"date"`
}

func (this *Download) AddRecord(dl *Download) bool {
	myorm.BSQL().Insert(TABLE_NAME_DOWNLOAD).Values("version", "channel", "type",
		"account", "city", "date")

	n, err := myorm.Insert(dl.Version, dl.Channel, dl.Type, dl.Account,
		dl.City, dl.Date)
	if err != nil {
		log.Error(err)
		return false
	}
	if n > 0 {
		return true
	}
	return false
}

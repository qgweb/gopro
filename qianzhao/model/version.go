package model

import (
	"github.com/ngaut/log"
	"strings"
)

const (
	VERSION_TABLE_NAME = "221su_version"
)

type Version struct {
	Id             int    `json:"id"`
	Version        string `json:"version"`
	Type           string `json:type"`
	Url            string `json:url"`
	Date           int    `json:date"`
	Download_count int    `json:download_count"`
	Update_page    string `json:update_page"`
}

type VersionExt struct {
	Version
	IsUpdate bool
}

func (this *Version) Update(v string) VersionExt {
	//获取渠道
	datas := strings.Split(v, ".")
	qudao := datas[len(datas)-1]
	vs := VersionExt{}
	myorm.BSQL().Select("*").From(VERSION_TABLE_NAME).Where("type=?").Order("date desc").Limit(1)
	list, err := myorm.Query(qudao)
	if err != nil {
		log.Warn("[version upload] 查询失败,", err)
		return vs
	}

	if len(list) > 0 && list[0]["version"] != v {
		vs.Update_page = list[0]["update_page"]
		vs.Url = list[0]["url"]
		vs.IsUpdate = true
	}

	return vs
}

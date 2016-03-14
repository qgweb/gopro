package model

import (
	"github.com/ngaut/log"
)

//"github.com/qgweb/gopro/qianzhao/common/function"

//"github.com/qgweb/gopro/lib/convert"

const (
	FAVORITE_TABLE_NAME = "221su_favorite"
)

type Favorite struct {
	Uid     string `uid`
	Content string `content`
}

// 获取收藏夹
func (this *Favorite) GetFavorite(uid string) (f Favorite) {
	sql := myorm.BSQL().Select("*").From(FAVORITE_TABLE_NAME).Where("uid=?").GetSQL()
	list, err := myorm.Query(sql, uid)
	if err != nil {
		log.Warn("[favorite getfavorite] 读取数据出错", err)
		return
	}

	if len(list) == 0 {
		return
	}

	f.Uid = list[0]["uid"]
	f.Content = list[0]["content"]
	return
}

// 保存收藏夹
func (this *Favorite) SaveFavorite(f Favorite) bool {
	// 是否存在
	sql := myorm.BSQL().Select("count(*) as num").From(FAVORITE_TABLE_NAME).Where("uid=?").GetSQL()
	list, err := myorm.Query(sql, f.Uid)
	if err != nil {
		log.Warn("[favorite saveFavorite] 查找失败 ", err)
		return false
	}

	if len(list) > 0 && list[0]["num"] == "0" {
		sql := myorm.BSQL().Insert(FAVORITE_TABLE_NAME).Values("uid", "content").GetSQL()
		n, err := myorm.Insert(sql, f.Uid, f.Content)
		if err != nil {
			log.Warn("[favorite saveFavorite] 插入失败 ", err)
			return false
		}

		if n > 0 {
			return true
		}

		return false
	} else {
		sql := myorm.BSQL().Update(FAVORITE_TABLE_NAME).Set("content").Where("uid=?").GetSQL()
		n, err := myorm.Update(sql, f.Content, f.Uid)
		if err != nil {
			log.Warn("[favorite saveFavorite] 保存失败 ", err)
			return false
		}

		if n > 0 {
			return true
		}

		return false
	}
}

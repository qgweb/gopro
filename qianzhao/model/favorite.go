package model

import

//"github.com/goweb/gopro/qianzhao/common/function"

"log"

//"github.com/goweb/gopro/lib/convert"

const (
	TABLE_NAME = "221su_favorite"
)

type Favorite struct {
	Uid     string `uid`
	Content string `content`
}

// 获取收藏夹
func (this *Favorite) GetFavorite(uid string) (f Favorite) {
	myorm.BSQL().Select("*").From(TABLE_NAME).Where("uid=?")
	list, err := myorm.Query(uid)
	if err != nil {
		log.Println("[favorite getfavorite] 读取数据出错", err)
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
	myorm.BSQL().Select("count(*) as num").From(TABLE_NAME).Where("uid=?")
	list, err := myorm.Query(f.Uid)
	if err != nil {
		log.Println("[favorite saveFavorite] 查找失败 ", err)
		return false
	}

	if len(list) > 0 && list[0]["num"] == "0" {
		myorm.BSQL().Insert(TABLE_NAME).Values("uid", "content")
		n, err := myorm.Insert(f.Uid, f.Content)
		if err != nil {
			log.Println("[favorite saveFavorite] 插入失败 ", err)
			return false
		}

		if n > 0 {
			return true
		}

		return false
	} else {
		myorm.BSQL().Update(TABLE_NAME).Set("content").Where("uid=?")
		n, err := myorm.Update(f.Content, f.Uid)
		if err != nil {
			log.Println("[favorite saveFavorite] 保存失败 ", err)
			return false
		}

		if n > 0 {
			return true
		}

		return false
	}
}

package model

import

//"github.com/goweb/gopro/qianzhao/common/function"

"log"

//"github.com/goweb/gopro/lib/convert"

type Favorite struct {
	Uid     string `uid`
	Content string `content`
}

// 获取收藏夹
func (this *Favorite) GetFavorite(uid string) (f Favorite) {
	myorm.BSQL().Select("*").From("221su_favorite").Where("uid=?")
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

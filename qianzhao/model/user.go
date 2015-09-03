package model

import (
	"fmt"
	"strings"

	"github.com/goweb/gopro/qianzhao/common/function"

	"log"

	"github.com/goweb/gopro/lib/convert"
)

type User struct {
	Id            string `id`
	Name          string `username`
	Pwd           string `password`
	Avatar        string `avatar`
	Bandwith      string `bandwith`
	BandwithPwd   string `bandwith_pwd`
	AppUid        string `app_uid`
	Created       int    `created`
	RememberToken string `remember_token`
	Sid           string `sid`
}

// 用户名是否存在
func (this *User) UserNameExist(name string) bool {
	myorm.BSQL().Select("count(*) as num").From("221su_users").Where("username=?")
	list, err := myorm.Query(name)
	if err != nil {
		log.Println("[user UserNameExist]数据获取失败", err)
		return false
	}

	if len(list) > 0 && list[0]["num"] == "1" {
		return true
	}

	return false
}

// 用户是否存在
func (this *User) UserExist(name string, pwd string) bool {
	ui := this.UserInfo(name)
	if ui.Name == "" {
		return false
	}

	return function.CheckBcrypt([]byte(ui.Pwd), []byte(pwd))
}

// 用户信息
func (this *User) UserInfo(name string) (u User) {
	myorm.BSQL().Select("*").From("221su_users").Where("username=?")
	list, err := myorm.Query(name)
	if err != nil {
		log.Println("[user UserInfo]数据获取失败", err)
		return
	}

	if len(list) == 0 {
		return
	}

	u.Id = list[0]["id"]
	u.Name = list[0]["username"]
	u.Pwd = list[0]["password"]
	u.Avatar = list[0]["avator"]
	u.Bandwith = list[0]["bandwith"]
	u.BandwithPwd = list[0]["bandwith_pwd"]
	u.AppUid = list[0]["app_uid"]
	u.Created = convert.ToInt(list[0]["created"])
	u.RememberToken = list[0]["remember_token"]
	u.Sid = list[0]["sid"]

	return u
}

// 用户注册
func (this *User) UserRegister(name string, password string) bool {
	myorm.BSQL().Insert("221su_users").Values("username", "password", "created")
	n, err := myorm.Insert(name, function.GetBcrypt([]byte(password)), function.GetTimeUnix())
	if err != nil {
		log.Println("[user UserRegister] 插入失败，", err)
		return false
	}

	if n > 0 {
		return true
	}

	return false
}

// 更新
func (this *User) Update(values map[string]interface{}, wheres map[string]interface{}) bool {
	var (
		fields = make([]string, 0, len(values))
		where  = make([]string, 0, len(wheres))
		wvlues = make([]interface{}, 0, len(wheres))
	)

	for k, v := range values {
		fields = append(fields, k)
		wvlues = append(wvlues, v)
	}

	for k, v := range wheres {
		where = append(where, k+"=?")
		wvlues = append(wvlues, v)
	}

	fmt.Println(myorm.BSQL().Update("221su_users").Set(fields...).Where(strings.Join(where, " and ")))
	n, err := myorm.Update(wvlues...)
	if err != nil {
		log.Println("[user Update]更新失败", err)
		return false
	}
	if n > 0 {
		return true
	}

	return false
}

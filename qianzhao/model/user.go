package model

import (
	"github.com/ngaut/log"
	"strings"

	"github.com/qgweb/gopro/lib/encrypt"

	"github.com/qgweb/gopro/qianzhao/common/function"

	"github.com/qgweb/gopro/lib/convert"
)

const (
	USER_TABLE_NAME = "221su_users"
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
	Email         string `email`
	Phone         string `phone`
}

func make_app_uid(bandwith string, bandwith_pwd string, timestamp string) string {
	return encrypt.DefaultMd5.Encode(bandwith + bandwith_pwd + timestamp + "1024")
}

// 用户名是否存在
func (this *User) UserNameExist(name string) bool {
	sql := myorm.BSQL().Select("count(*) as num").From(USER_TABLE_NAME).Where("username=? or phone=? or email=?").GetSQL()
	list, err := myorm.Query(sql, name, name, name)
	if err != nil {
		log.Warn("[user UserNameExist]数据获取失败", err)
		return false
	}
	log.Info(name)
	if len(list) > 0 && list[0]["num"] == "1" {
		return true
	}

	return false
}

// 判断手机是否存在
func (this *User) PhoneExist(phone string) bool {
	sql := myorm.BSQL().Select("count(*) as num").From(USER_TABLE_NAME).Where("phone=?").GetSQL()
	list, err := myorm.Query(sql, phone)
	if err != nil {
		log.Warn("[user UserInfo]数据获取失败", err)
		return false
	}
	if len(list) == 0 {
		return false
	}

	if convert.ToInt(list[0]["num"]) > 0 {
		return true
	}

	return false
}

// 判断邮箱是否存在
func (this *User) EmailExist(email string) bool {
	sql := myorm.BSQL().Select("count(*) as num").From(USER_TABLE_NAME).Where("email=?").GetSQL()
	list, err := myorm.Query(sql, email)
	if err != nil {
		log.Warn("[user UserInfo]数据获取失败", err)
		return false
	}
	if len(list) == 0 {
		return false
	}

	if convert.ToInt(list[0]["num"]) > 0 {
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
	sql := myorm.BSQL().Select("*").From(USER_TABLE_NAME).Where("username=? or phone=? or email=?").GetSQL()
	list, err := myorm.Query(sql, name, name, name)
	if err != nil {
		log.Warn("[user UserInfo]数据获取失败", err)
		return
	}

	if len(list) == 0 {
		return
	}

	u.Id = list[0]["id"]
	u.Name = list[0]["username"]
	u.Pwd = list[0]["password"]
	u.Avatar = list[0]["avatar"]
	u.Bandwith = list[0]["bandwith"]
	u.BandwithPwd = list[0]["bandwith_pwd"]
	u.AppUid = list[0]["app_uid"]
	u.Created = convert.ToInt(list[0]["created"])
	u.RememberToken = list[0]["remember_token"]
	u.Sid = list[0]["sid"]
	u.Email = list[0]["email"]
	u.Phone = list[0]["phone"]

	return u
}

// 用户信息
func (this *User) UserInfoById(id string) (u User) {
	sql := myorm.BSQL().Select("*").From(USER_TABLE_NAME).Where("id=?").GetSQL()
	list, err := myorm.Query(sql, id)
	if err != nil {
		log.Warn("[user UserInfo]数据获取失败", err)
		return
	}

	if len(list) == 0 {
		return
	}

	u.Id = list[0]["id"]
	u.Name = list[0]["username"]
	u.Pwd = list[0]["password"]
	u.Avatar = list[0]["avatar"]
	u.Bandwith = list[0]["bandwith"]
	u.BandwithPwd = list[0]["bandwith_pwd"]
	u.AppUid = list[0]["app_uid"]
	u.Created = convert.ToInt(list[0]["created"])
	u.RememberToken = list[0]["remember_token"]
	u.Sid = list[0]["sid"]
	u.Email = list[0]["email"]
	u.Phone = list[0]["phone"]

	return u
}

// 用户注册
func (this *User) UserRegister(phone string, email string, password string) bool {
	uname := "qz_" + function.GetTimeUnix()
	sql := myorm.BSQL().Insert(USER_TABLE_NAME).Values("phone", "email", "password", "created", "username").GetSQL()
	n, err := myorm.Insert(sql, phone, email, function.GetBcrypt([]byte(password)), function.GetTimeUnix(), uname)
	if err != nil {
		log.Warn("[user UserRegister] 插入失败，", err)
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

	sql := myorm.BSQL().Update(USER_TABLE_NAME).Set(fields...).Where(strings.Join(where, " and ")).GetSQL()
	n, err := myorm.Update(sql, wvlues...)
	if err != nil {
		log.Warn("[user Update]更新失败", err)
		return false
	}
	if n > 0 {
		return true
	}

	return false
}

// 验证宽带
func (this *User) VerifyBandWith(bandwith string, bandwith_pwd string) (app_uid string) {
	sql := myorm.BSQL().Select("app_uid").From(USER_TABLE_NAME).Where("bandwith=?").GetSQL()
	list, err := myorm.Query(sql, bandwith)
	if err != nil {
		log.Warn("[user VerifyBandWith]查询失败", err)
		return
	}

	if len(list) > 0 {
		return list[0]["app_uid"]
	}

	timestamp := function.GetTimeUnix()
	app_uid = make_app_uid(bandwith, bandwith_pwd, timestamp)

	sql = myorm.BSQL().Insert(USER_TABLE_NAME).Values("bandwith", "bandwith_pwd", "created", "app_uid").GetSQL()
	n, err := myorm.Insert(sql, bandwith, bandwith_pwd, timestamp, app_uid)
	if err != nil {
		log.Warn("[user VerifyBandWith] 插入失败", err)
		return ""
	}

	if n > 0 {
		return
	}

	return ""
}

// 获取宽带
func (this *User) GetBrandWith(app_uid string) (u User) {
	sql := myorm.BSQL().Select("bandwith", "id").From(USER_TABLE_NAME).Where("app_uid=?").GetSQL()
	list, err := myorm.Query(sql, app_uid)
	if err != nil {
		log.Warn("[user GetBrandWith] 查询失败", err)
		return
	}

	if len(list) > 0 {
		u.Bandwith = list[0]["bandwith"]
		u.Id = list[0]["id"]
		return u
	}
	return
}

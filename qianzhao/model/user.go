package model

import (
	"errors"
	"fmt"
	"github.com/ngaut/log"
	"github.com/qgweb/gopro/lib/convert"
	"github.com/qgweb/gopro/lib/encrypt"
	"github.com/qgweb/gopro/qianzhao/common/function"
	"strings"
	"time"
)

const (
	USER_TABLE_NAME = "221su_users"
)

var ERR_Award_NOT_ALLOW = errors.New("非校园用户")

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
	AwardCount    int    `award_count`
}

func make_app_uid(bandwith string, bandwith_pwd string, timestamp string) string {
	return encrypt.DefaultMd5.Encode(bandwith + bandwith_pwd + timestamp + "1024")
}

func (this *User) To(info map[string]string) (u User) {
	u.Id = info["id"]
	u.Name = info["username"]
	u.Pwd = info["password"]
	u.Avatar = info["avatar"]
	u.Bandwith = info["bandwith"]
	u.BandwithPwd = info["bandwith_pwd"]
	u.AppUid = info["app_uid"]
	u.Created = convert.ToInt(info["created"])
	u.RememberToken = info["remember_token"]
	u.Sid = info["sid"]
	u.Email = info["email"]
	u.Phone = info["phone"]
	u.AwardCount = convert.ToInt(info["award_count"])
	return
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
	info, err := myorm.Get(sql, name, name, name)
	if err != nil {
		log.Warn("[user UserInfo]数据获取失败", err)
		return
	}

	return this.To(info)
}

// 用户信息
func (this *User) GetUserIdByPhone(phone string) (u User) {
	sql := myorm.BSQL().Select("*").From(USER_TABLE_NAME).Where("phone=?").GetSQL()
	info, err := myorm.Get(sql, phone)
	if err != nil {
		log.Warn("[user UserInfo]数据获取失败", err)
		return
	}

	return this.To(info)
}

// 用户信息
func (this *User) UserInfoById(id string) (u User) {
	sql := myorm.BSQL().Select("*").From(USER_TABLE_NAME).Where("id=?").GetSQL()
	info, err := myorm.Get(sql, id)
	if err != nil {
		log.Warn("[user UserInfo]数据获取失败", err)
		return
	}
	return this.To(info)
}

// 获取抽奖次数
func (this *User) getAwardCount() string {
	if time.Since(time.Date(2016, 9, 1, 0, 0, 0, 0, time.Local)).Seconds() > 0 {
		return "1"
	}
	return "0"
}

// 用户注册
func (this *User) UserRegister(phone string, email string, password string) bool {
	uname := "qz_" + function.GetTimeUnix()
	sql := myorm.BSQL().Insert(USER_TABLE_NAME).Values("phone", "email", "password", "created", "username", "award_count").GetSQL()
	n, err := myorm.Insert(sql, phone, email, function.GetBcrypt([]byte(password)), function.GetTimeUnix(), uname, this.getAwardCount())
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

// 获取可抽奖次数
func (this *User) GetAwardCount(uid string) (int, error) {
	u := this.UserInfoById(uid)
	return u.AwardCount, nil
}

// 添加或删减抽奖次数
func (this *User) IncrAwardCount(uid int, count int) (bool, error) {
	sql := fmt.Sprintf("update 221su_users set award_count = award_count +%d where id=%d", count, uid)
	fmt.Println(sql)
	n, err := myorm.Update(sql)
	return n > 0, err
}

// 判断是否可参与活动
func (this *User) CanAward(phone string) (bool, error) {
	sql := myorm.BSQL().Select("account").From("221su_broadband").Where("account=?").GetSQL()
	info, err := myorm.Get(sql, phone)
	if err != nil {
		return false, err
	}
	if len(info) > 0 {
		return true, nil
	}
	return false, ERR_Award_NOT_ALLOW
}

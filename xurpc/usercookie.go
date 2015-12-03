//用户标签
package main

import (
	"errors"
	"fmt"
	"github.com/qgweb/gopro/lib/encrypt"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type UserCookieData struct {
}

//获取用户标签
func (this UserCookieData) GetUserTags(cid string) []byte {
	user_cox_ua, err := getUserCoxUa(cid)
	if err != nil || len(user_cox_ua) == 0 {
		return jsonReturn(nil, err)
	}
	md5 := getMd5(user_cox_ua)
	tags, err := getTags(md5)
	if err != nil {
		return jsonReturn(nil, err)
	}
	return jsonReturn(tags, err)
}

/**
 * 根据cookie查询用户的cox和ua
 */
func getUserCoxUa(id string) (map[string]interface{}, error) {
	var (
		db     = IniFile.Section("mongo-cookie").Key("db").String()
		table  = "dt_user"
		sess   = getcookieSession()
		result map[string]interface{}
	)
	defer sess.Close()

	if !bson.IsObjectIdHex(id) {
		return nil, errors.New("请输入mongoId")
	}
	err := sess.DB(db).C(table).FindId(bson.ObjectIdHex(id)).One(&result)
	if mgo.ErrNotFound != err && err != nil {
		return nil, errors.New("无数据或查询有误")
	}
	return result, nil
}

/**
 * 获取cox和ua的md5值
 */
func getMd5(result map[string]interface{}) string {
	md5 := encrypt.DefaultMd5.Encode(result["cox"].(string) +
		encrypt.DefaultBase64.Encode(result["ua"].(string)))
	return md5
}

/**
 * 获取用户标签
 */
func getTags(md5 string) ([]string, error) {
	var (
		db     = IniFile.Section("mongo-data_source").Key("db").String()
		table  = "useraction_temp_tags"
		sess   = getcattjSession()
		result map[string]interface{}
		tags   []string
	)
	defer sess.Close()
	err := sess.DB(db).C(table).Find(bson.M{"adua": md5}).One(&result)

	if mgo.ErrNotFound != err && err != nil {
		return nil, err
	}
	if len(result) > 0 {
		for _, v := range result["tag"].([]interface{}) {
			tags = append(tags, v.(string))
		}
	}
	return tags, nil
}

//获取session
func getcookieSession() *mgo.Session {
	var (
		mouser = IniFile.Section("mongo-cookie").Key("user").String()
		mopwd  = IniFile.Section("mongo-cookie").Key("pwd").String()
		mohost = IniFile.Section("mongo-cookie").Key("host").String()
		moport = IniFile.Section("mongo-cookie").Key("port").String()
		modb   = IniFile.Section("mongo-cookie").Key("db").String()
		url    = fmt.Sprintf("%s:%s/%s", mohost, moport, modb)
	)
	if mouser != "" && modb != "" {
		url = fmt.Sprintf("%s:%s@%s", mouser, mopwd, url)
	}

	mdbsession, err := mgo.Dial(url)
	if err != nil {
		panic(err)
	}
	return mdbsession

}

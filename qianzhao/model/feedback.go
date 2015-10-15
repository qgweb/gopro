package model

import (
	"errors"
	"github.com/ngaut/log"
	"reflect"
	"time"
)

type FeedBack struct {
	Id        int    `json:"id"`        //编号
	Btype     int    `json:"type"`      //浏览器版本
	Qtype     int    `json:"qtype"`     //问题分类
	QDescribe string `json:"qdescribe"` //问题描述
	Qpic      string `json:"qpic"`      //图片
	Contact   string `json:"contact"`   //联系方式
	Tcontact  int    `json:"tcontact"`  //联系方式类型
	Status    int    `json:"status"`    //状态
	Time      int    `json:"time"`      //时间

}

const (
	TABLE_NAME_FEEDBACK = "221su_feedback"
)

// 添加记录
func (this *FeedBack) AddRecord(fb *FeedBack) bool {
	myorm.BSQL().Insert(TABLE_NAME_FEEDBACK).Values(getFields(FeedBack{})...)
	n, err := myorm.Insert(fb.Btype, fb.Qtype, fb.QDescribe, fb.Qpic, fb.Contact, fb.Tcontact, 0, time.Now().Unix())
	if err != nil {
		log.Error(err)
		return false
	}
	if n > 0 {
		return true
	}
	return false
}

// 验证数据有效性
func (this *FeedBack) CheckData(fb *FeedBack) error {
	if fb.Btype == 0 {
		return errors.New("浏览器版本不能为空")
	}
	if fb.Qtype == 0 {
		return errors.New("问题分类不能为空")
	}
	if fb.QDescribe == "" {
		return errors.New("问题描述不能为空")
	}
	if fb.Tcontact == 0 {
		return errors.New("联系方式类型不能为空")
	}
	if fb.Contact == "" {
		return errors.New("联系方式不能为空")
	}
	return nil
}

// 获取字段
func getFields(f interface{}) []string {
	ref := reflect.TypeOf(f)
	num := ref.NumField()
	vs := make([]string, 0, num)
	for i := 0; i < num; i++ {
		tag := ref.Field(i).Tag.Get("json")
		if tag == "id" {
			continue
		}
		vs = append(vs, tag)
	}
	return vs
}

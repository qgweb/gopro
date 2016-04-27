package controllers

import (
	"fmt"
	"github.com/astaxie/beego"
	"github.com/qgweb/gopro/shrtb_houtai/models"
	"strings"
)

type OrderController struct {
	BaseController
}

// 订单列表
func (this *OrderController) List() {
	order := models.Order{}
	list, err := order.List()
	if err != nil {
		beego.Error(err)
	}
	this.Data["pageTitle"] = "订单列表"
	this.Data["list"] = list
	this.Data["pageBar"] = ""
	this.display()
}

// 解析订单url
func (this *OrderController) parsePutUrls(value string) map[string]interface{} {
	var list = make(map[string]interface{}, 100)
	for _, v := range strings.Split(value, "\n") {
		list[strings.TrimSpace(v)] = 1
	}
	return list
}

// 添加订单
func (this *OrderController) Add() {
	if this.isPost() {
		o := models.Order{}
		if err := this.ParseForm(&o); err != nil {
			this.ajaxMsg("参数解析错误", MSG_ERR)
		}
		o.Purl = this.parsePutUrls(this.Input().Get("purl"))
		if err :=o.Add(o);err !=nil {
			this.ajaxMsg(fmt.Sprint(o), MSG_ERR)
		}
		this.ajaxMsg("添加成功", MSG_OK)
	}

	this.Data["pageTitle"] = "添加订单"
	this.display()
}

// 删除订单
func (this *OrderController) Del() {
	name :=this.Input().Get("name")
	o:=models.Order{}
	o.Del(name)
	this.redirect("/order/list")
}

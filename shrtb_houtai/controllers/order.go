package controllers

import (
	"strings"

	"github.com/astaxie/beego"
	"github.com/qgweb/gopro/shrtb_houtai/models"
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
		if strings.TrimSpace(v) == "" {
			continue
		}
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
		if err := o.Add(o); err != nil {
			this.ajaxMsg("添加失败", MSG_ERR)
		}
		this.ajaxMsg("添加成功", MSG_OK)
	}

	this.Data["pageTitle"] = "添加订单"
	this.LayoutSections = make(map[string]string)
	this.LayoutSections["ProTime"] = "include/time.html"
	this.display()
}

// 删除订单
func (this *OrderController) Del() {
	var o models.Order
	id := this.GetString("id")
	beego.Info(id)
	if err := o.Del(id); err == nil {
		this.ajaxMsg("删除成功", MSG_OK)
	} else {
		this.ajaxMsg("删除失败"+err.Error(), MSG_ERR)
	}
}

// 编辑订单
func (this *OrderController) Edit() {
	var (
		order = models.Order{}
		id    = this.GetString("id")
	)

	if this.isPost() {
		o := models.Order{}
		if err := this.ParseForm(&o); err != nil {
			this.ajaxMsg("参数解析错误", MSG_ERR)
		}
		o.Stats = "未投放"
		if err := o.Edit(o); err != nil {
			this.ajaxMsg("修改失败"+err.Error(), MSG_ERR)
		}
		this.ajaxMsg("修改成功", MSG_OK)
	}

	if strings.TrimSpace(id) == "" {
		this.redirect("/order/list")
	}
	if info, err := order.GetId(id); err != nil {
		this.redirect("/order/list")
	} else {
		if strings.TrimSpace(info.TotalLimit) == "9999999" {
			info.TotalLimit = ""
		}
		if strings.TrimSpace(info.DayLimit) == "9999999" {
			info.DayLimit = ""
		}
		this.Data["info"] = info
	}
	this.Data["pageTitle"] = "编辑订单"
	this.LayoutSections = make(map[string]string)
	this.LayoutSections["ProTime"] = "include/time.html"
	this.display()
}

func (this *OrderController) Open() {
	var (
		oid    = this.Input().Get("oid")
		status = this.Input().Get("stats")
		o      models.Order
	)

	if oid == "" || status == "" {
		this.ajaxMsg("参数错误", MSG_ERR)
	}
	if err := o.Push(oid, status); err != nil {
		this.ajaxMsg("修改失败", MSG_ERR)
	} else {
		this.ajaxMsg("修改成功", MSG_OK)
	}
}

func (this *OrderController) Report() {
	return
	o := &models.Order{}
	o.Putin()
	this.Ctx.Output.JSON("ok", true, true)
}

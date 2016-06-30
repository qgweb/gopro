package controllers

import (
	"github.com/astaxie/beego"
	"github.com/qgweb/gopro/shrtb_houtai/models"
)

type ReportController struct {
	BaseController
}

// 订单列表
func (this *ReportController) List() {
	report := models.Report{}
	list, err := report.List()
	if err != nil {
		beego.Error(err)
	}
	this.Data["pageTitle"] = "报表列表"
	this.Data["list"] = list
	this.Data["pageBar"] = ""
	this.display()
}

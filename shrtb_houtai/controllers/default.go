package controllers

import (
	"github.com/astaxie/beego"
	"strings"
	"github.com/lisijie/webcron/app/libs"
	"strconv"
)

type MainController struct {
	BaseController
}

func (c *MainController) Get() {
	c.Data["Website"] = "beego.me"
	c.Data["Email"] = "astaxie@gmail.com"
	c.TplName = "index.tpl"
}

func (this *MainController) List() {
	this.Ctx.WriteString("ok")
}

func (this *MainController) Login() {
	this.Data["siteName"] = "上海投放订单系统"
	if this.userId > 0 {
		this.redirect("/")
	}
	beego.ReadFromRequest(&this.Controller)
	if this.isPost() {
		flash := beego.NewFlash()

		username := strings.TrimSpace(this.GetString("username"))
		password := strings.TrimSpace(this.GetString("password"))
		remember := this.GetString("remember")
		if username != "" && password != "" {
			errorMsg := ""
			if username != "qgshrtb" || password != "qgshrtb123" {
				errorMsg = "帐号或密码错误"
			} else {
				authkey := libs.Md5([]byte(this.getClientIp() + "|qgshrtb123"))
				if remember == "yes" {
					this.Ctx.SetCookie("auth", strconv.Itoa(1)+"|"+authkey, 7*86400)
				} else {
					this.Ctx.SetCookie("auth", strconv.Itoa(1)+"|"+authkey)
				}
				this.redirect(beego.URLFor("MainController.List"))
			}
			flash.Error(errorMsg)
			flash.Store(&this.Controller)
			this.redirect(beego.URLFor("MainController.Login"))
		}
	}

	this.TplName = "main/login.html"
}
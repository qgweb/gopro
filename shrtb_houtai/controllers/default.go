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
	c.redirect("/login")
}

func (this *MainController) List() {
	this.Ctx.WriteString("ok")
}

func (this *MainController) Login() {
	if this.userId > 0 {
		this.redirect("/order/list")
	}
	beego.ReadFromRequest(&this.Controller)
	if this.isPost() {
		flash := beego.NewFlash()
		username := strings.TrimSpace(this.GetString("username"))
		password := strings.TrimSpace(this.GetString("password"))
		remember := this.GetString("remember")
		if username != "" && password != "" {
			errorMsg := ""
			uname :=  beego.AppConfig.String("user::name")
			upwd := beego.AppConfig.String("user::pwd")
			if username != uname || password != upwd {
				errorMsg = "帐号或密码错误"
			} else {

				authkey := libs.Md5([]byte(this.getClientIp() + "|"+upwd))
				if remember == "yes" {
					this.Ctx.SetCookie("auth", strconv.Itoa(1)+"|"+authkey, 7*86400)
				} else {
					this.Ctx.SetCookie("auth", strconv.Itoa(1)+"|"+authkey)
				}
				this.redirect(beego.URLFor("OrderController.List"))
			}
			flash.Error(errorMsg)
			flash.Store(&this.Controller)
			this.redirect(beego.URLFor("MainController.Login"))
		}
	}

	this.TplName = "main/login.html"
}

// 退出登录
func (this *MainController) Logout() {
	this.Ctx.SetCookie("auth", "")
	this.redirect(beego.URLFor("MainController.Login"))
}
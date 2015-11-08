package controllers

import (
	"github.com/astaxie/beego"
	"github.com/ngaut/log"
	"github.com/qgweb/gopro/lib/ssh"
)

type MainController struct {
	beego.Controller
}

func (c *MainController) Get() {
	conf := ssh.Config{}
	conf.PrivaryKey = ssh.GetPrivateKey()
	conf.RemoteHost = "122.225.98.69"
	conf.RemoteUser = "root"
	l, err := ssh.NewSSHLinker(conf)
	if err != nil {
		c.Ctx.WriteString(err.Error())
		return
	}
	s, err := l.GetClient().NewSession()
	if err != nil {
		log.Error(err)
		return
	}
	bb, err := s.Output("curl 'http://192.168.0.72:4151/stats'")
	if err != nil {
		log.Error(err)
		return
	}
	c.Ctx.WriteString(string(bb))
}

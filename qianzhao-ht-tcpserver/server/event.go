// 事件服务器
package server

import (
	"encoding/json"
	"net"
	"strings"
	"time"

	"github.com/ngaut/log"

	"github.com/qgweb/gopro/lib/convert"
	"github.com/qgweb/gopro/qianzhao-ht-tcpserver/common/function"
	"github.com/qgweb/gopro/qianzhao-ht-tcpserver/model"
)

type Event struct{}

// 开启加速
func (this *Event) Start(conn *net.TCPConn, req *Request) {
	var (
		bd = &BDInterfaceManager{}
		ip = strings.Split(conn.RemoteAddr().String(), ":")[0]
	)

	log.Info(bd.FreeCardApplyFor("15158117079"))
	return
	account := bd.CanStart(ip)
	if account == "" {
		rep := Respond{}
		rep.Code = "500"
		rep.Msg = "抱歉，您的运行环境不符合加速条件"
		data, _ := MRespond(&rep)
		conn.Write(ProtocolPack(data))
		return
	}

	// 判断是否还有时间
	accountTime := bd.HaveTime(account)
	log.Warn(accountTime)
	if accountTime <= 0 {
		rep := Respond{}
		rep.Code = "500"
		rep.Msg = "抱歉，您的加速体验时间已结束"
		data, _ := MRespond(&rep)
		conn.Write(ProtocolPack(data))
		return
	}

	rep := bd.Start(account, ip)

	if rep.Code != "200" {
		data, _ := MRespond(&rep)
		conn.Write(ProtocolPack(data))
		return
	}

	connManager.Add(account, conn)
	user := &Account{}
	user.Name = account
	user.BTime = time.Now().Unix()
	user.RemoteAddr = ip
	user.ChannelId = rep.Msg
	user.CTime = int64(accountTime)
	accountManager.Add(user)

	// 返回数据
	data, _ := MRespond(&Respond{Code: "200", Msg: account + "|" + convert.ToString(accountTime)})
	conn.Write(ProtocolPack(data))
}

// 停止加速
func (this *Event) Stop(conn *net.TCPConn, req *Request) {
	var (
		bd      = &BDInterfaceManager{}
		account = req.Content
	)

	user := accountManager.Get(account)
	if user == nil {
		return
	}

	rep := bd.Stop(user.ChannelId)
	if rep.Code != "200" {
		data, _ := MRespond(&rep)
		conn.Write(ProtocolPack(data))
		return
	}

	//添加记录
	record := &model.BrandAccountRecord{}
	record.Account = user.Name
	record.BeginTime = user.BTime
	record.EndTime = time.Now().Unix()
	record.Date = convert.ToInt64(function.GetDateUnix())
	record.AddRecord(*record)

	//删除链接
	nreq := &Request{}
	nreq.Action = "stop"
	nreq.Content = ""
	data, _ := MRequest(nreq)
	conn.Write(ProtocolPack(data))

	accountManager.Del(account)
	connManager.Del(account)

}

// 内部调用停止
func (this *Event) InnerStop(account string) {
	conn := connManager.Get(account)
	if conn != nil {
		r := &Request{}
		r.Action = "stop"
		r.Content = account
		this.Stop(conn, r)
	}
}

// 响应ping
func (this *Event) RepPing(conn *net.TCPConn) {
	r := &Respond{}
	r.Code = "200"
	r.Msg = "ok"

	b, _ := MRespond(r)
	conn.Write(b)
}

// 检测内部程序状态
func (this *Event) Info(conn *net.TCPConn) {
	type Info struct {
		Conn     int
		Accounts []Account
	}

	info := Info{}
	info.Conn = connManager.Count()
	accounts := accountManager.Range()
	info.Accounts = make([]Account, 0, len(accounts))

	for _, v := range accounts {
		info.Accounts = append(info.Accounts, *v)
	}

	b, err := json.Marshal(&info)
	if err == nil {
		conn.Write(b)
	}
}

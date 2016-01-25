// 事件服务器
package server

import (
	"encoding/json"
	"net"
	"strings"
	"time"

	//"github.com/ngaut/log"

	"fmt"
	"github.com/ngaut/log"
	"github.com/qgweb/gopro/lib/convert"
	"github.com/qgweb/gopro/qianzhao-ht-tcpserver/common/function"
	"github.com/qgweb/gopro/qianzhao-ht-tcpserver/model"
)

type Event struct{}

// 连接管理
func (this *Event) Link(conn *net.TCPConn, req *Request) {
	var (
		ip      = strings.Split(conn.RemoteAddr().String(), ":")[0]
		content = req.Content
		resp    = Respond{}
		repAry  = strings.Split(content, "|")
	)

	if len(repAry) < 5 {
		resp.Code = "500"
		resp.Msg = "参数错误"
		data, _ := MRespond(&resp)
		conn.Write(ProtocolPack(data))
		return
	}

	connManager.Add(repAry[0], conn)
	user := &Account{}
	user.Name = repAry[0]
	user.BTime = time.Now().Unix()
	user.RemoteAddr = ip
	user.CTime = convert.ToInt64(repAry[4])
	accountManager.Add(user)

	resp.Code = "200"
	resp.Msg = "ok"
	data, _ := MRespond(&resp)
	conn.Write(ProtocolPack(data))
	return
}

// 开启加速
func (this *Event) Start(conn *net.TCPConn, req *Request) {
	var (
		bd     = &BDInterfaceManager{}
		reqAry = strings.Split(req.Content, "|")
		resp   Respond
	)

	if reqAry[0] == "0" { //免费卡
		mc := MCard{}
		mc.Mobile = reqAry[1]
		resp = bd.Start(mc, 0)
	}

	log.Error(reqAry)
	if reqAry[0] == "1" { //申请卡
		if len(reqAry) < 4 {
			data, _ := MRespond(&Respond{Code: "500", Msg: "参数错误"})
			conn.Write(ProtocolPack(data))
			return
		}
		mc := MCard{}
		mc.Mobile = reqAry[1]
		mc.CardNO = reqAry[2]
		mc.CardPass = reqAry[3]
		mc.Serviceid = "0001"
		resp = bd.Start(mc, 1)
	}

	if resp.Code != "200" {
		data, _ := MRespond(&resp)
		conn.Write(ProtocolPack(data))
		return
	}

	// 返回数据
	data, _ := MRespond(&resp)
	conn.Write(ProtocolPack(data))
	conn.Close()
}

// 停止加速
func (this *Event) Stop(conn *net.TCPConn, req *Request) {
	var (
		account = req.Content
	)

	user := accountManager.Get(account)
	if user == nil {
		return
	}

	//添加记录
	record := model.HtCardRecord{}
	record.HtId = convert.ToInt64(account)
	record.BeginTime = user.BTime
	record.EndTime = time.Now().Unix()
	record.Date = convert.ToInt64(function.GetDateUnix())
	record.AddRecord(record)

	//删除链接
	nreq := &Request{}
	nreq.Action = "stop"
	nreq.Content = ""
	data, _ := MRequest(nreq)
	conn.Write(ProtocolPack(data))
	conn.Close()
	accountManager.Del(account)
	connManager.Del(account)
}

// 内部调用停止
func (this *Event) InnerStop(account string) {
	log.Info(account)
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

func (this *Event) HaveCardByPhone(conn *net.TCPConn, req *Request) {
	var (
		phone  = req.Content
		hmodel = model.HTCard{}
	)

	if ht := hmodel.GetMoneyLastCard(phone); ht.Id > 0 {
		r := &Respond{}
		r.Code = "199"
		r.Msg = fmt.Sprintf("%s|%s", ht.CardNum, ht.CardPwd)

		b, _ := MRespond(r)
		conn.Write(ProtocolPack(b))
	} else {
		r := &Respond{}
		r.Code = "500"
		r.Msg = "无绑定"
		b, _ := MRespond(r)
		conn.Write(ProtocolPack(b))
	}
	conn.Close()
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

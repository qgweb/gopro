// 事件服务器
package server

import (
	"encoding/json"
	"github.com/qgweb/gopro/lib/convert"
	"github.com/qgweb/gopro/qianzhao-tcpserver/model"
	"log"
	"net"
	"strings"
	"time"
)

type Event struct{}

// 开启加速
func (this *Event) Start(conn *net.TCPConn, req *Request) {
	var (
		bd      = &BDInterfaceManager{}
		account = req.Content
		ip      = conn.RemoteAddr().String()
		//ip = "121.237.226.1:11137"
	)

	rep := bd.CanStart(account, ip)
	if rep.Code != "200" {
		data, _ := MRespond(&rep)
		conn.Write(ProtocolPack(data))
		return
	}

	//rep.Msg = "120|0002"
	udata := strings.Split(rep.Msg, "|") // "时长|区域"

	// 把socket连接加入连接管理
	connManager.Add(account, conn)
	// 把用户加入用户管理
	user := Account{}
	user.Name = account
	user.BTime = time.Now().Unix()
	user.CTime = convert.ToInt64(udata[0])
	user.RemoteAddr = ip
	user.Area = udata[1]
	accountManager.Add(&user)

	data, _ := MRespond(&Respond{Code: "200", Msg: "ok"})
	conn.Write(ProtocolPack(data))
}

// 停止加速
func (this *Event) Stop(conn *net.TCPConn, req *Request) {
	var (
		bd      = &BDInterfaceManager{}
		account = req.Content
		user    = accountManager.Get(account)
		ip      = conn.RemoteAddr().String()
		//ip = "121.237.226.1:11137"
	)

	// 用户不能存在返回
	if user == nil {
		return
	}

	rep := bd.Stop(account, user.Area, ip)
	if rep.Code != "200" {
		data, _ := MRespond(&rep)
		conn.Write(ProtocolPack(data))
		return
	}

	user.ETime = time.Now().Unix()

	// 把记录添加到数据库
	record := &model.BrandAccountRecord{}
	record.Account = account
	record.BeginTime = user.BTime
	record.EndTime = user.ETime
	record.Time = time.Now().Unix()
	record.AddRecord(*record)

	// 用户可使用时间修改
	brand := &model.BrandAccount{}
	ba, err := brand.GetAccountInfo(account)
	if err == nil && ba.Id != "" {
		ba.UsedTime = ba.UsedTime + int(record.EndTime-record.BeginTime)
		if ba.UsedTime > ba.TotalTime {
			ba.UsedTime = ba.TotalTime
		}
		log.Println(ba)
		brand.EditAccount(ba)
	}
	// 发送关闭成功
	nreq := &Request{}
	nreq.Action = "stop"
	nreq.Content = ""
	data, _ := MRequest(nreq)
	conn.Write(ProtocolPack(data))
	// 删除
	accountManager.Del(account)
	// 去掉socket连接
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
	log.Println(len(accounts))
	info.Accounts = make([]Account, 0, len(accounts))

	for _, v := range accounts {
		info.Accounts = append(info.Accounts, *v)
	}

	b, err := json.Marshal(&info)
	if err == nil {
		conn.Write(b)
	}
}

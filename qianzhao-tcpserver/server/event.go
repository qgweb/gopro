// 事件服务器
package server

import (
	"net"
)

var (
	ConnManager = NewAccountConnManager()
)

type Event struct{}

// 开启加速
func (this *Event) Start(conn *net.TCPConn, req *Request) {

}

// 停止加速
func (this *Event) Stop(conn *net.TCPConn, req *Request) {

}

// 响应ping
func (this *Event) RepPing(conn *net.TCPConn) {
	r := &Respond{}
	r.Code = "200"
	r.Msg = "ok"

	b, _ := MRespond(r)
	conn.Write(b)
}

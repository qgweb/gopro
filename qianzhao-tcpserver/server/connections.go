package server

import (
	"encoding/json"
	"net"
	"time"

	"sync"
)

type ConnectionManager struct {
	*sync.WaitGroup
	Counter int
}

func NewConnectionManager() *ConnectionManager {
	cm := &ConnectionManager{}
	cm.WaitGroup = &sync.WaitGroup{}
	return cm
}

func (cm *ConnectionManager) Add(delta int) {
	cm.Counter += delta
	cm.WaitGroup.Add(delta)
}

func (cm *ConnectionManager) Done() {
	cm.Counter--
	cm.WaitGroup.Done()
}

// 账户连接管理
type AccountConnManager struct {
	sync.Mutex
	conns map[string]*net.TCPConn
}

// 新建管理
func NewAccountConnManager() *AccountConnManager {
	return &AccountConnManager{conns: make(map[string]*net.TCPConn)}
}

// 添加连接
func (this *AccountConnManager) Add(name string, conn *net.TCPConn) {
	this.Lock()
	defer this.Unlock()
	this.conns[name] = conn
}

// 删除连接
func (this *AccountConnManager) Del(name string) {
	this.Lock()
	defer this.Unlock()
	if v, ok := this.conns[name]; ok {
		v.Close()
		delete(this.conns, name)
	}
}

// 获取连接
func (this *AccountConnManager) Get(name string) *net.TCPConn {
	this.Lock()
	defer this.Unlock()
	return this.conns[name]
}

// 返回连接数目
func (this *AccountConnManager) Count() int {
	this.Lock()
	defer this.Unlock()
	return len(this.conns)
}

// ping服务
func (this *AccountConnManager) Ping(fun func(name string)) {
	t := time.NewTicker(septime)
	for {
		select {
		case <-t.C:
			for k, conn := range this.conns {
				r := &Request{}
				r.Action = "ping"
				b, _ := MRequest(r)
				conn.Write(ProtocolPack(b))
				buf := make([]byte, 100)
				conn.SetReadDeadline(time.Now().Add(time.Second*5))
				_, err := conn.Read(buf)
				if err != nil {
					// 客户端已挂掉
					fun(k)
					continue
				}
			}
		}
	}
}

// 请求
type Request struct {
	Action  string `json:"action"`
	Content string `json:"content"`
}

// 响应
type Respond struct {
	Code string `json:"code"`
	Msg  string `json:"msg"`
}

// 解析请求
func UmRequest(data []byte) (Request, error) {
	r := Request{}
	err := json.Unmarshal(data, &r)
	if err != nil {
		return r, err
	}
	return r, nil
}

// 序列化请求
func MRequest(req *Request) ([]byte, error) {
	d, err := json.Marshal(req)
	if err != nil {
		return d, err
	}
	return d, nil
}

// 解析响应
func UnRespond(data []byte) (Request, error) {
	r := Request{}
	err := json.Unmarshal(data, &r)
	if err != nil {
		return r, err
	}
	return r, nil
}

// 序列化请求
func MRespond(req *Respond) ([]byte, error) {
	d, err := json.Marshal(req)
	if err != nil {
		return d, err
	}
	return d, nil
}

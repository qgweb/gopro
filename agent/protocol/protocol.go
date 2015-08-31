package protocol

import (
	"bytes"
	"fmt"
	"github.com/goweb/gopro/lib/encrypt"
	"net"
	"sync"
	"time"
)

const (
	PROTOCOL_HEAD    = "QGWEB"
	APP_STATUS_STOP  = "STOP"
	APP_STATUS_RUN   = "RUNNING"
	APP_STATUS_ERROR = "ERROR"
)

type PeerInfo struct {
	Hostname      string `json:"hostname"`
	RemoteAddress string `json:"remote_address"`
	ProgrameName  string `json:"programe_name"`
	Status        string `json:"status"`
	Version       string `json:"version"`
	LastUpdate    int64  `json:"last_update"`
}

type Programe struct {
	PeerInfo
	Conn *net.TCPConn
}

type RegistrationDB struct {
	sync.RWMutex
	registrationMap map[string]*Programe
}

// 初始化
func NewRegistrationDB() (p *RegistrationDB) {
	p = &RegistrationDB{}
	p.registrationMap = make(map[string]*Programe)
	return
}

// 注册
func (this *RegistrationDB) Register(info *Programe) {
	this.Lock()
	key := encrypt.DefaultMd5.Encode(info.Hostname + info.RemoteAddress)
	this.registrationMap[key] = info
	this.Unlock()
}

// 移除
func (this *RegistrationDB) UnRegister(key string) {
	this.Lock()
	delete(this.registrationMap, key)
	this.Unlock()
}

// 遍历
func (this *RegistrationDB) Range() (info map[string]*Programe) {
	this.Lock()
	// 防止主数据被污染
	for k, v := range this.registrationMap {
		info[k] = v
	}
	this.Unlock()
	return
}

// 更新状态
func (this *RegistrationDB) UpdateStatus(key string, status string) {
	this.Lock()
	this.registrationMap[key].Status = status
	this.Unlock()
}

// 更新时间
func (this *RegistrationDB) UpdateLastTime(key string) {
	this.Lock()
	this.registrationMap[key].LastUpdate = time.Now().Unix()
	this.Unlock()
}

// 状态检测
func (this *RegistrationDB) HealthCheck() {
	// 一分钟自检
	t := time.NewTicker(time.Minute)
	for {
		select {
		case <-t.C:
			fmt.Println("ok")
			for k, v := range this.registrationMap {
				if v.Conn == nil {
					this.Lock()
					this.registrationMap[k].Status = APP_STATUS_ERROR
					this.Unlock()
					continue
				}

				fmt.Println(k, v.Hostname, v.RemoteAddress, v.Status)
				//检查时间是否超过2分钟
				if time.Now().Sub(time.Unix(v.LastUpdate, 0)).Minutes() >= 1 {
					this.Lock()
					this.registrationMap[k].Status = APP_STATUS_ERROR
					this.Unlock()

					//发送心跳包, 检查程序是否还存在
					res := this.checkHealth(this.registrationMap[k].Conn)
					if res == false {
						this.Lock()
						this.registrationMap[k].Status = APP_STATUS_STOP
						this.Unlock()
					} else {
						this.Lock()
						this.registrationMap[k].Status = APP_STATUS_RUN
						this.Unlock()
					}
				}
				fmt.Println(k, v.Hostname, v.RemoteAddress, v.Status)
			}
		}
	}
}

// 验证是否存在
func (this *RegistrationDB) checkHealth(conn *net.TCPConn) bool {
	_, err := conn.Write(ProtocolPack([]byte("HEART")))
	if err != nil {
		return false
	}
	buf := make([]byte, 20)
	n, err := conn.Read(buf)
	if bytes.Equal(ProtocolUnPack(buf[0:n]), []byte("OK")) {
		return true
	}

	return false
}

// 封包
func ProtocolPack(data []byte) []byte {
	p := NewProtocol(PROTOCOL_HEAD)
	return p.Packet(data)
}

// 解包
func ProtocolUnPack(data []byte) []byte {
	p := NewProtocol(PROTOCOL_HEAD)
	b, _ := p.Unpack(data)
	return b[0]
}

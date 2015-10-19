package main

import (
	"sync"
	"time"
)

// 主机信息
type Hosts struct {
	Name   string   `json:"name"`
	Ip     string   `json:"ip"`
	Pid    string   `json:"pid"` //区分同一个ip多台机器
	Info   InfoData `json:"infodata"`
	Uptime int64    `json:"time"` // 更新时间
}

// 反馈的数据
type InfoData struct {
	Type       string `json:"type"` //类型
	ReceiveNum int    `json:"rnum"` //接收数据
	DealNum    int    `json:"dnum"` //处理数据
}

type HostsManager struct {
	sync.RWMutex
	data map[string]*Hosts
}

func NewHostsManager() *HostsManager {
	return &HostsManager{
		data: make(map[string]*Hosts),
	}
}

// 添加记录
func (this *HostsManager) Add(h *Hosts) {
	this.Lock()
	defer this.Unlock()

	key := h.Ip + "_" + h.Pid
	if v, ok := this.data[key]; ok {
		v.Uptime = time.Now().Unix()
		v.Info = h.Info
	} else {
		this.data[key] = h
	}
}

// 获取所有记录
func (this *HostsManager) Range() []Hosts {
	this.RLock()
	defer this.RUnlock()

	list := make([]Hosts, 0, len(this.data))

	for _, v := range this.data {
		list = append(list, *v)
	}

	return list
}

// 去掉挂掉的机器
func (this *HostsManager) Run() {
	t := time.NewTicker(time.Minute)
	for {
		select {
		case <-t.C:
			this.Lock()
			for k, v := range this.data {
				if time.Now().Unix()-v.Uptime >= 30 {
					delete(this.data, k)
				}
			}
			this.Unlock()
		}
	}
}

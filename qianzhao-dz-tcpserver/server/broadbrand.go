// 宽带服务管理
package server

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"sync"
	"time"

	"github.com/qgweb/gopro/qianzhao-dz-tcpserver/logger"
)

var (
	septime = time.Second * 5
)

// 宽带账号管理
type AccountManager struct {
	sync.Mutex
	StopChan     chan bool
	StopDiskChan chan bool
	OverChan     chan bool
	Users        map[string]*Account // 账号集合
	log          *logger.Logger
}

type Account struct {
	Name       string `json:"account"`     // 账号名称
	BTime      int64  `json:"begin_time"`  // 开始时间
	ETime      int64  `json:"end_time"`    // 结束时间
	CTime      int64  `json:"used_time"`   // 可使用时间
	ChannelId  string `json:"channel_id:`  // 加速频道ID
	RemoteAddr string `json:"remote_addr"` // 客户端的地址
}

// 新建管理服务
func NewAccountManager(log *logger.Logger) *AccountManager {
	return &AccountManager{
		Users:        make(map[string]*Account),
		log:          log,
		StopChan:     make(chan bool),
		StopDiskChan: make(chan bool),
		OverChan:     make(chan bool, 2),
	}
}

// 添加用户
func (this *AccountManager) Add(user *Account) {
	this.Lock()
	defer this.Unlock()
	this.Users[user.Name] = user
}

// 删除用户
func (this *AccountManager) Del(name string) {
	this.Lock()
	defer this.Unlock()
	delete(this.Users, name)
}

// 获取用户
func (this *AccountManager) Get(name string) *Account {
	this.Lock()
	defer this.Unlock()
	if u, ok := this.Users[name]; ok {
		return u
	}
	return nil
}

// 编辑用户
func (this *AccountManager) Edit(user *Account) {
	this.Lock()
	defer this.Unlock()
	this.Users[user.Name] = user
}

// 遍历
func (this *AccountManager) Range() map[string]*Account {
	this.Lock()
	defer this.Unlock()
	return this.Users
}

// 定时刷新到磁盘
func (this *AccountManager) TimeFlushDisk(fileName string) {
	var flushDiskFun = func() {
		this.Lock()
		defer this.Unlock()

		bytes, err := json.Marshal(&this.Users)
		if err != nil {
			this.log.Println("json格式化失败,错误信息为:", err)
			return
		}

		err = ioutil.WriteFile(fileName, bytes, os.ModePerm)
		if err != nil {
			this.log.Println("更新到磁盘失败,错误信息为:", err)
			return
		}
	}

	t := time.NewTicker(septime)
	for {
		select {
		case <-t.C: // 检测时间
			flushDiskFun()
		case <-this.StopDiskChan:
			this.log.Println("接收到磁盘停止信号")
			// 后续处理
			flushDiskFun()
			this.log.Println("定时磁盘服务已接收到停止服务")
			close(this.StopDiskChan)
			this.OverChan <- true
			return
		}
	}
}

// 定时检测时间
func (this *AccountManager) TimeCheckAccountUTime(fun func(name string)) {
	t := time.NewTicker(septime)
	for {
		select {
		case <-t.C: // 检测时间
			ntime := time.Now().Unix()
			for k, v := range this.Range() {
				this.log.Println(k)
				// 定期把使用时间写入
				u := this.Get(k)
				u.ETime = ntime
				this.Edit(u)

				// 检测是否到时
				if ntime-v.BTime >= v.CTime {
					v.ETime = ntime
					fun(k) //回调函数处理
					// 删除用户
					this.Del(k)
				}
			}
		case <-this.StopChan:
			// 后续处理
			this.log.Println("接收到停止信号")
			for k, _ := range this.Range() {
				// 定期把使用时间写入
				u := this.Get(k)
				u.ETime = time.Now().Unix()
				this.Edit(u)
				// 回调
				fun(k)
			}

			this.log.Println("定时时钟服务已接收到停止服务")
			close(this.StopChan)
			this.OverChan <- true
			return
		}
	}
}

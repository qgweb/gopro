package main

import (
	"container/list"
	"runtime"
	"sync"
	"time"

	"github.com/ngaut/log"
)

type Goods map[string]interface{}

type DockerNode struct {
	InputPipe  chan string //输入管道
	OutputPipe chan Goods  //输出管道
	Size       int         //管道大小
	NoticeChan chan bool   // 通知信号
}

type GrabFactory struct {
	InputPipe      chan *DockerNode // 输入管道
	PipeSize       int              // 管道缓存大小
	buffer         *list.List       //缓冲池
	ConcurrentSize int              //发发执行大小
	sync.Mutex
}

func NewFactory(size int, csize int) *GrabFactory {
	return &GrabFactory{
		InputPipe:      make(chan *DockerNode, size),
		PipeSize:       size,
		ConcurrentSize: csize,
		buffer:         list.New()}
}

func (this *GrabFactory) Add(node *DockerNode) {
	this.InputPipe <- node
}

func (this *GrabFactory) Push() {
	for {
		v := <-this.InputPipe
		log.Warn(this.buffer.Len(), len(this.InputPipe))
		this.Mutex.Lock()
		this.buffer.PushBack(v)
		this.Mutex.Unlock()
		runtime.Gosched()
	}
}

func (this *GrabFactory) Grab(fun func(string) Goods) {
	this.Mutex.Lock()
	defer this.Mutex.Unlock()
	node := this.buffer.Front()
	tmpeSize := 0

	if node == nil {
		return
	}

	// 删除已处理过的数据
	for node != nil {
		nd := node.Value.(*DockerNode)
		if len(nd.OutputPipe) == nd.Size {
			log.Error("notice")
			nd.NoticeChan <- true
			close(nd.OutputPipe)
			tmp := node
			this.buffer.Remove(tmp)
		}

		node = node.Next()
	}

	node = this.buffer.Front()
	for node != nil {
		if tmpeSize = tmpeSize + len(node.Value.(*DockerNode).InputPipe); tmpeSize >= this.ConcurrentSize {
			goto LABEL
		}
		node = node.Next()
	}

	return
LABEL:
	head := this.buffer.Front()
	tmpeSize = 0
	wg := sync.WaitGroup{}
	defer func() {
		log.Warn("waitwait")
		wg.Wait()
		time.Sleep(time.Second * 3)
	}()
	for head != nil {
		nd := head.Value.(*DockerNode)
		iplen := len(nd.InputPipe)
		if iplen <= this.ConcurrentSize-tmpeSize {
			for i := 0; i < iplen; i++ {
				wg.Add(1)
				go func(v string, c chan Goods) {
					//log.Info(v)
					c <- fun(v)
					wg.Done()
				}(<-nd.InputPipe, nd.OutputPipe)
			}
			tmpeSize = tmpeSize + iplen
		} else {
			for i := 0; i < this.ConcurrentSize-tmpeSize; i++ {
				wg.Add(1)
				go func(v string, c chan Goods) {
					//log.Info(v)
					c <- fun(v)
					wg.Done()
				}(<-nd.InputPipe, nd.OutputPipe)
			}
			tmpeSize = this.ConcurrentSize
			return
		}
		head = head.Next()
	}
}

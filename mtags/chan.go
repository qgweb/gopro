package main

import (
	"github.com/ngaut/log"
	"math"
	"sync"
)

type ChanFunction struct {
	Fun    func(...interface{})
	Params []interface{}
}

// 并发执行程序
type ChanFactory struct {
	chanSize int // 池大小
	pageSize int // 每页大小
	funcPool []ChanFunction
}

func NewChanFactory(psize int, csize int) *ChanFactory {
	return &ChanFactory{pageSize: psize, chanSize: csize,
		funcPool: make([]ChanFunction, 0, csize)}
}

func (this *ChanFactory) Push(fun ChanFunction) {
	this.funcPool = append(this.funcPool, fun)
}

func (this *ChanFactory) Run() {
	pageCount := int(math.Ceil(float64(this.chanSize) / float64(this.pageSize)))
	wg := sync.WaitGroup{}

	for p := 1; p <= pageCount; p++ {
		begin := (p - 1) * this.pageSize
		end := begin + this.pageSize
		if p == pageCount {
			end = this.chanSize
		}

		for _, v := range this.funcPool[begin:end] {
			wg.Add(1)
			go func(p []interface{}) {
				defer func() {
					if msg := recover(); msg != nil {
						log.Error(msg)
					}
				}()

				v.Fun(p...)
				wg.Done()
			}(v.Params)
		}
		wg.Wait()
	}
}

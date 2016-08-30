// 生产者和消费者
package Concurency

import (
	"github.com/ngaut/log"
	"runtime/debug"
)

type ConsumerFun func(int, interface{}) error

// 并发读取
type ProducterConsumer struct {
	productChan chan interface{}
	over        chan struct{}
}

func NewConcurrencyGet(proSize int, cumSize int) *ProducterConsumer {
	return &ProducterConsumer{
		make(chan interface{}, proSize),
		make(chan struct{}, cumSize),
	}
}

func (this *ProducterConsumer) CloseProduct() {
	close(this.productChan)
}

func (this *ProducterConsumer) Push(v interface{}) {
	this.productChan <- v
}

func (this *ProducterConsumer) Pop(fun ConsumerFun) {
	for i := 0; i < cap(this.over); i++ {
		go func(index int) {
			for {
				v, ok := <-this.productChan
				if !ok {
					this.over <- struct{}{}
					break
				}
				func() {
					defer func() {
						if msg := recover(); msg != nil {
							log.Error(msg)
							debug.PrintStack()
						}
					}()
					log.Error(fun(index, v))
				}()
			}
		}(i)
	}
}

func (this *ProducterConsumer) Wait() {
	for i := 0; i < cap(this.over); i++ {
		<-this.over
	}
}

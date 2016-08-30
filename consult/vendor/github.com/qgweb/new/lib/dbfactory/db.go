package dbfactory

import (
	"sync"
)

type DataInSourceFun func() chan interface{}
type DataInStorageFun func(out interface{})
type DataOutSourceFun func() chan interface{}
type DataOutStorageFun func(out interface{})

type Factory struct {
	inSourcePool   []DataInSourceFun
	outStoragePool []DataOutStorageFun
}

func New(insize, outsize int) *Factory {
	f := &Factory{}
	f.inSourcePool = make([]DataInSourceFun, 0, insize)
	f.outStoragePool = make([]DataOutStorageFun, 0, outsize)
	return f
}

func (this *Factory) AddInSource(fun DataInSourceFun) {
	this.inSourcePool = append(this.inSourcePool, fun)
}

func (this *Factory) AddOutSource(fun DataOutStorageFun) {
	this.outStoragePool = append(this.outStoragePool, fun)
}

func (this *Factory) InRun(fff DataInStorageFun) {
	if len(this.inSourcePool) == 0 {
		return
	}
	wg := sync.WaitGroup{}
	for _, v := range this.inSourcePool {
		wg.Add(1)
		go func(f DataInSourceFun) {
			var res = f()
			for {
				select {
				case v, ok := <-res:
					if ok {
						fff(v)
					} else {
						wg.Done()
						goto END
					}
				}
			}
		END:
		}(v)
	}
	wg.Wait()
}

func (this *Factory) OutRun(fff DataOutSourceFun) {
	if len(this.outStoragePool) == 0 {
		return
	}

	var res = fff()
	for {
		select {
		case v, ok := <-res:
			if ok {
				wg := sync.WaitGroup{}
				for _, vv := range this.outStoragePool {
					wg.Add(1)
					go func(f DataOutStorageFun) {
						defer wg.Done()
						f(v)
					}(vv)
					wg.Wait()
				}
			} else {
				goto END
			}
		}
	}
END:
}

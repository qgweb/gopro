package main

import (
	"github.com/qgweb/gopro/lib/grab"
	"sync"
	"fmt"
	"time"
	"github.com/qgweb/gopro/lib/convert"
)

func main() {
	wg := sync.WaitGroup{}
	bt:= time.Now()
	msgs:= make(chan string, 100)
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			h:=grab.GrabTaoHTML("http://www.taobao.com")
			msgs <- convert.ToString(len(h))
			wg.Done()
		}()
	}
	wg.Wait()
	close(msgs)
	fmt.Println(time.Now().Sub(bt).Seconds())
	for k := range msgs {
		fmt.Println(k)
	}

}

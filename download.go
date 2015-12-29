package main

import (
	"github.com/ngaut/log"
	"github.com/qgweb/gopro/lib/mulgrab"
	"sync"
	"time"
)

func main() {
	config := mulgrab.Config{}
	config.DeadlineTimeOut = time.Minute
	config.DialTimeout = time.Minute
	grab1 := mulgrab.New(config)

	wg := sync.WaitGroup{}
	bt := time.Now()
	contents := make(chan string, 4)

	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			//http.Get("https://www.taobao.com")
			v, err := grab1.Get("https://www.taobao.com", nil)
			if err != nil {
				contents <- ""
			} else {
				contents <- string(v[0:1])
			}
			wg.Done()
		}()
	}

	wg.Wait()
	close(contents)
	log.Info(time.Now().Sub(bt).Seconds())
	bt = time.Now()
	//grab1.Get("http://www.taobao.com", nil)
	log.Info(time.Now().Sub(bt).Seconds())
	for v := range contents {
		log.Info(len(v))
	}
}

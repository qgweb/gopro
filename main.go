package main

import (
	"flag"
	"fmt"
	"github.com/garyburd/redigo/redis"
	"log"
	"sync"
	"time"
)

func newPool(server, password string) *redis.Pool {
	return &redis.Pool{
		MaxIdle:     50,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", server)
			if err != nil {
				return nil, err
			}
			if password != "" {
				if _, err := c.Do("AUTH", password); err != nil {
					c.Close()
					return nil, err
				}
			}
			c.Do("SELECT", 6)
			return c, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
}

func main() {
	var (
		host     = flag.String("host", "192.168.1.199", "域名")
		port     = flag.String("port", "6380", "端口")
		poolchan = make(chan int, 50)
		exitchan = make(chan bool)
		size     = 100000
		wg       = sync.WaitGroup{}
	)

	flag.Parse()

	pool := newPool(fmt.Sprintf("%s:%s", *host, *port), "")
	bt := time.Now()
	go func() {
		for i := 0; i < size; i++ {
			poolchan <- i
		}
		close(poolchan)
	}()

	go func() {
		for v := range poolchan {
			wg.Add(1)
			go func(vv int) {
				conn := pool.Get()
				conn.Do("SET", vv, vv)
				conn.Close()
				wg.Done()
			}(v)
		}
		wg.Wait()
		exitchan <- true
	}()

	<-exitchan
	log.Println(time.Now().Sub(bt).Seconds())
	pool.Close()
}

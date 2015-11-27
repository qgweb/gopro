package main

import (
	"github.com/garyburd/redigo/redis"
	"log"
	"time"
)

func main() {
	conn, _ := redis.Dial("tcp4", "127.0.0.1:6379")
	defer conn.Close()
	bt := time.Now()

	for i := 0; i < 100000; i++ {
		conn.Do("SET", i, i)
	}

	log.Println(time.Now().Sub(bt).Seconds())
}

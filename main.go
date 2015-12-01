package main

import (
	"flag"
	"github.com/garyburd/redigo/redis"
	"log"
	"time"
)

func main() {
	var (
		host = flag.String("host", "192.168.0.75", "域名")
		port = flag.String("port", "6379", "端口")
	)

	flag.Parse()

	conn, _ := redis.Dial("tcp4", *host+":"+*port)
	defer conn.Close()
	conn.Do("SELECT", "6")
	bt := time.Now()
	for i := 0; i < 100000; i++ {
		conn.Do("SET", i, i)
	}

	log.Println(time.Now().Sub(bt).Seconds())
	conn.Do("FLUSHDB")
}

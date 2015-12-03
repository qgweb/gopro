package main

import (
	"github.com/qgweb/gopro/lib/convert"
	"github.com/seefan/gossdb"
	"log"
	"time"
)

func main() {
	pool, err := gossdb.NewPool(&gossdb.Config{
		Host:             "127.0.0.1",
		Port:             8888,
		MinPoolSize:      5,
		MaxPoolSize:      50,
		AcquireIncrement: 5,
	})
	if err != nil {
		log.Fatal(err)
		return
	}

	c, err := pool.NewClient()
	if err != nil {
		log.Println(err.Error())
		return
	}
	defer c.Close()
	bt := time.Now()
	for i := 0; i < 10000; i++ {
		c.Set(convert.ToString(i), i)
	}

	log.Println(time.Now().Sub(bt).Seconds())

}

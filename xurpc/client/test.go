package main

import (
	"github.com/hprose/hprose-go"
	"fmt"
	"time"
)

type Func struct {
	GetTagCount func(string, int) []byte
}

func main() {
	client := hprose.NewHttpClient("http://127.0.0.1:12345")
	var f Func
	client.UseService(&f)
	bt := time.Now()
	fmt.Println(string(f.GetTagCount("11111", 2)))
	fmt.Println(time.Since(bt).Seconds())
}

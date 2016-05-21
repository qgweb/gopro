package main

import (
	"github.com/hprose/hprose-go"
	"fmt"
	"time"
)

type Func struct {
	GetTagCount func(province string, tp string, tagid string, day int) []byte
}

func main() {
	client := hprose.NewHttpClient("http://192.168.1.199:12345")
	var f Func
	client.UseService(&f)
	bt := time.Now()
	fmt.Println(string(f.GetTagCount("zj","url","111",1)))
	fmt.Println(time.Since(bt).Seconds())
}

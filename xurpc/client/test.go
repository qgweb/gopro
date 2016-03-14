package main

import (
	"github.com/hprose/hprose-go"
	"fmt"
	"time"
)

type Func struct {
	GetShopCount func(string, string) int
}

func main() {
	client := hprose.NewHttpClient("http://127.0.0.1:12345")
	var f Func
	client.UseService(&f)
	bt := time.Now()
	fmt.Println(f.GetShopCount("112608855", "1455933600"))
	fmt.Println(time.Since(bt).Seconds())
}

package main

import (
	"github.com/hprose/hprose-go"
	"fmt"
)

type Func struct {
	StatusHandler func(string, string, string) []byte
}

func main() {
	client := hprose.NewHttpClient("http://127.0.0.1:12345")
	var f Func
	client.UseService(&f)
	fmt.Println(string(f.StatusHandler("13977", "strategy", "1")))
}

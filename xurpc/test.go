package main

import (
	"fmt"

	"github.com/hprose/hprose-go/hprose"
)

type clientStub struct {
	Hello      func(string) string
	Swap       func(int, int) (int, int)
	Sum        func(...int) (int, error)
	GetTaoData func(string) int
}

func main() {
	client := hprose.NewClient("http://192.168.1.199:12344/")
	var ro clientStub
	client.UseService(&ro)
	fmt.Println(ro.GetTaoData("World"))
}

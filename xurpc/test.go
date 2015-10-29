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
	GetTag     func(string, string) []byte
}

func main1() {
	client := hprose.NewClient("http://127.0.0.1:12345/")
	var ro clientStub
	client.UseService(&ro)
	fmt.Println(ro.GetTaoData("World"))
	fmt.Println(string(ro.GetTag("42591253308", "")))
}

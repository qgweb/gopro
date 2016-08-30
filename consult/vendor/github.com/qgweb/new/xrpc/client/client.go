package main

import (
	"github.com/hprose/hprose-go/hprose"
	"net/http"
)

type Test struct {
	DomainCookie        func(string, string) error `name:"domain-visitor"`
	XXX                 func(map[string]interface{})
	RecordAdvertPutInfo func(map[string]string) error `name:"reocrd-advert"`
}

var (
	client hprose.Client
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		client := hprose.NewClient("http://127.0.0.1:12345")
		var ro Test
		client.UseService(&ro)
		(ro.RecordAdvertPutInfo(map[string]string{
			"ad":     "aaaa",
			"ua":     "bbbb",
			"pv":     "1",
			"click":  "1",
			"advert": "22",
		}))
	})
	http.ListenAndServe(":8888", nil)
}

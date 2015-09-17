package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
)

func main1() {
	request, _ := http.NewRequest("GET", "https://www.taobao.com/", nil)
	proxy, err := url.Parse("http://192.168.1.122:8888")
	if err != nil {
		log.Println(err)
		return
	}
	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxy),
		},
	}
	res, err := client.Do(request)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer res.Body.Close()
	aa, _ := ioutil.ReadAll(res.Body)
	fmt.Println(string(aa))
}

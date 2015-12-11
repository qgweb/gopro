package main

import (
	"flag"
	"fmt"
	"github.com/astaxie/beego/httplib"
	// "github.com/donnie4w/dom4g"
	//"github.com/opesun/goquery"
)

const (
	QOS_QUERY_URL = "http://202.102.13.98:7001/services/LcimsForUserInfo"
)

func main() {
	var num = flag.String("n", "", "卡号")
	flag.Parse()
	var query = `<x:Envelope xmlns:x="http://schemas.xmlsoap.org/soap/envelope/">
  <x:Header/>
  <x:Body>
    <getCardUserInfo xmlns="">
      <cardno>` + *num + `</cardno>
    </getCardUserInfo>
  </x:Body>
</x:Envelope>`
	req := httplib.Post(QOS_QUERY_URL)
	req.Header("Content-Type", "text/xml; charset=utf-8")
	req.Header("SOAPAction", QOS_QUERY_URL)
	req.Body(query)
	fmt.Println(req.String())
}

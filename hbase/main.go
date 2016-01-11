package main

import (
	"github.com/ngaut/log"
	"github.com/pingcap/go-hbase"
	"github.com/qgweb/gopro/lib/convert"
	"time"
)

/**
{
"_id" : ObjectId("55c1be799633a885e5e2f764"),
"ad" : "YwdLb3Y5dEltBGl2aRhgSQ==",
"date" : "2015-08-05",
"ua" : "Mozilla/4.0 (compatible; MSIE 8.0; YYGameAll_1.2.161288.80; Windows NT 5.1; Trident/4.0; QQDownload 708; qdesk 2.3.1186.202; .NET CLR 2.0.50727; u9dnfsh)",
"cids" : {
"12" : "50008882"
}
}

{
"_id" : ObjectId("55c1be799633a885e5e2f765"),
"ad" : "YwdLb3Y5dEltBGl2aRhgSQ==",
"date" : "2015-08-05",
"ua" : "Mozilla/4.0 (compatible; MSIE 9.0; QDesk 2.3.1185.202; Windows NT 6.1; WOW64; Trident/4.0; BTRS31753; SLCC2; .NET CLR 2.0.50727; .NET CLR 3.5.30729; .NET CLR 3.0.30729; Media Center PC 6.0; .NET4.0C)",
"cids" : {
"12" : "50013045"
}
}

{
"_id" : ObjectId("55c1be799633a885e5e2f766"),
"ad" : "YwdLb3Y5dEltBGl2aRhgSQ==",
"date" : "2015-08-05",
"ua" : "Mozilla/4.0 (compatible; MSIE 9.0; Windows NT 6.1; .NET CLR 2.0.50727; .NET CLR 3.0.04506.648; MAXTHON)",
"cids" : {
"12" : "50012038"
}
}
*/

type ZKer interface {
	 Put(string)
}

type SK struct {
	Name string
}

func (this *SK) Put( a string) {
	log.Info(a)
}

var _ ZKer = (*SK)(nil)

func main() {

	var a = SK{}
	a.Put("111")


	client, err := hbase.NewClient([]string{"192.168.1.218:2181"}, "/hbase")
	if err != nil {
		log.Error(err)
		return
	}
	defer client.Close()

	bt:=time.Now()
	for i:=0; i < 1000;i++ {
		p := hbase.NewPut([]byte("ffffff" + convert.ToString(i)))
		p.AddStringValue("cf", "13", "dsf")
		client.Put("test", p)
	}

	log.Info(time.Now().Sub(bt).Seconds())
}

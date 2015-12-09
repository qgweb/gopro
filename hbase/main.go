package main

import (
	"fmt"
	"git.apache.org/thrift.git/lib/go/thrift"
	"github.com/qgweb/gopro/hbase/gen-go/hbase"
	"os"

	"net"
)

func main() {
	protocolFactory := thrift.NewTBinaryProtocolFactoryDefault()
	trans, err := thrift.NewTSocket(net.JoinHostPort("127.0.0.1", "9090"))
	if err != nil {
		fmt.Fprintln(os.Stderr, "error resolving address:", err)
		os.Exit(1)
	}
	defer trans.Close()
	fmt.Println(trans.Open())
	client := hbase.NewTHBaseServiceClientFactory(trans, protocolFactory)
	// arg63 := "x"
	// mbTrans64 := thrift.NewTMemoryBufferLen(len(arg63))
	// defer mbTrans64.Close()
	// mbTrans64.WriteString(arg63)
	// factory66 := thrift.NewTBinaryProtocolFactoryDefault()
	// jsProt67 := factory66.GetProtocol(mbTrans64)
	// argvalue1 := hbase.NewTGet()
	// fmt.Println(argvalue1.Read(jsProt67))
	fmt.Println(client.Exists([]byte("t1"), hbase.NewTGet()))
}

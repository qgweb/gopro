package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"

	"github.com/qgweb/gopro/qianzhao-ht-tcpserver/protocol"
)

const (
	PROTOCOL_HEAD = "qgbrower"
)

// 请求
type Request struct {
	Action  string `json:"action"`
	Content string `json:"content"`
}

// 响应
type Respond struct {
	Code string `json:"code"`
	Msg  string `json:"msg"`
}

// 解析请求
func UmRequest(data []byte) (Request, error) {
	r := Request{}
	err := json.Unmarshal(data, &r)
	if err != nil {
		return r, err
	}
	return r, nil
}

// 序列化请求
func MRequest(req *Request) ([]byte, error) {
	d, err := json.Marshal(req)
	if err != nil {
		return d, err
	}
	return d, nil
}

// 解析响应
func UnRespond(data []byte) (Request, error) {
	r := Request{}
	err := json.Unmarshal(data, &r)
	if err != nil {
		return r, err
	}
	return r, nil
}

// 序列化请求
func MRespond(req *Respond) ([]byte, error) {
	d, err := json.Marshal(req)
	if err != nil {
		return d, err
	}
	return d, nil
}

// 封包
func ProtocolPack(data []byte) []byte {
	p := protocol.NewProtocol(PROTOCOL_HEAD)
	return p.Packet(data)
}

// 解包
func ProtocolUnPack(data []byte) []byte {
	p := protocol.NewProtocol(PROTOCOL_HEAD)
	b, err := p.Unpack(data)
	fmt.Println(string(err))
	return b[0]
}

func main() {

	//addr, _ := net.ResolveTCPAddr("tcp4", "122.225.98.80:9092")
	addr, _ := net.ResolveTCPAddr("tcp4", "122.225.98.80:9092")
	conn, err := net.DialTCP("tcp4", nil, addr)
	if err != nil {
		log.Fatalln("fuck")
	}

	r := &Request{}
	r.Action = "link"
	r.Content = ""
	d, _ := MRequest(r)
	conn.Write(ProtocolPack(d))
	buf := make([]byte, 2048)

	for {
		n, err := conn.Read(buf)
		if err != nil {
			log.Println(err)
		}

		log.Println(buf[0:n])
		log.Println(string(ProtocolUnPack(buf[0:n])))
		// time.Sleep(time.Second * 30)

		// r.Action = "stop"
		// r.Content = "10327158471"
		// d, _ := MRequest(r)
		// conn.Write(ProtocolPack(d))
		// n, err = conn.Read(buf)
		// if err != nil {
		// 	log.Println(err)
		// }

		// log.Println(string(ProtocolUnPack(buf[0:n])))
	}

}

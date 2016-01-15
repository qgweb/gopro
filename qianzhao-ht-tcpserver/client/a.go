package main

import (
	"encoding/json"
	"log"
	"net"

	"flag"
	"github.com/qgweb/gopro/qianzhao-ht-tcpserver/protocol"
	"io"
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
	b, _ := p.Unpack(data)
	return b[0]
}

func main() {
	var action = flag.String("a", "0", "操作:0 免费卡, 1 申请卡,2 验证是否有卡")
	var server = flag.String("s", "0", "服务器")
	var host = "127.0.0.1"
	flag.Parse()
	//qianzhao.221su.com
	if *server == "1" {
		host = "qianzhao.221su.com"
	}
	addr, _ := net.ResolveTCPAddr("tcp4", host+":9093")
	conn, err := net.DialTCP("tcp4", nil, addr)
	if err != nil {
		log.Fatalln("fuck")
	}

	r := &Request{}
	switch *action {
	case "0":
		r.Action = "plink"
		r.Content = "0|15158117079"
	case "1":
		r.Action = "plink" //56000005039489          WhVgF3cR //56000005040361          SRgZTmDu
		r.Content = "1|15158117079|56000005040361|SRgZTmDu"
	case "2":
		r.Action = "havebind"
		r.Content = "56000005038843|zta3t7M0"
	case "3":
		r.Action = "link"
		r.Content = "36|56000005040361|SRgZTmDu|1449817500610|7194"
	case "4":
		r.Action = "info"
		r.Content = ""
	}

	d, _ := MRequest(r)

	buf := make([]byte, 2048)
	// log.Println(ProtocolPack(d))

	conn.Write(ProtocolPack(d))

	for {
		n, err := conn.Read(buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Println(err)
		}

		log.Println(string(buf[0:n]))
		log.Println(len(buf[0:n]))
	}

}

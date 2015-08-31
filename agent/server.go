package main

import (
	"bytes"
	"log"
	"net"
	"time"

	"./protocol"
)

var (
	regDB *protocol.RegistrationDB
)

func init() {
	regDB = protocol.NewRegistrationDB()
}

func Handle(conn *net.TCPConn, param []byte) {
	res := bytes.Split(param, []byte(" "))
	switch {
	case bytes.Equal(res[0], []byte("REGISTER")):
		p := &protocol.Programe{}
		ress := bytes.Split(res[1], []byte("\t"))
		p.Hostname = string(ress[0])
		p.RemoteAddress = conn.RemoteAddr().String()
		p.ProgrameName = string(ress[1])
		p.LastUpdate = time.Now().Unix()
		p.Status = protocol.APP_STATUS_RUN
		p.Conn = conn
		regDB.Register(p)
		log.Println(regDB)
	}

	conn.Write([]byte("ok"))

}

func Show(conn *net.TCPConn) {
	for {
		buf := make([]byte, 1024)
		n, err := conn.Read(buf)
		if err != nil {
			log.Println(err)
			conn.Close()
			break
		}

		Handle(conn, protocol.ProtocolUnPack(buf[0:n]))
	}
}

func main() {
	l, err := net.Listen("tcp", "127.0.0.1:3000")
	if err != nil {
		log.Fatalln("监听端口失败")
	}

	go regDB.HealthCheck()

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Println("接收失败")
			continue
		}
		go Show(conn.(*net.TCPConn))
	}

}

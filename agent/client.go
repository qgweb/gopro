// package main

// import (
// 	"./protocol"
// 	"log"
// 	"net"
// 	"os"
// )

// func main() {
// 	t, err := net.ResolveTCPAddr("tcp", "127.0.0.1:3000")
// 	if err != nil {
// 		log.Fatalln(err)
// 	}
// 	conn, err := net.DialTCP("tcp", nil, t)
// 	if err != nil {
// 		log.Fatalln(err)
// 	}
// 	name, _ := os.Hostname()
// 	conn.Write(protocol.ProtocolPack([]byte("REGISTER " + name + "\t" + "TCPSERVER")))
// 	buf := make([]byte, 1024)
// 	n, err := conn.Read(buf)
// 	if err != nil {
// 		conn.Close()
// 		return
// 	}
// 	log.Println(string(buf[0:n]))
// 	conn.Close()
// }

package main

import (
	"fmt"
)

func main() {
	for {
		var str string
		fmt.Scanf("%s\n", &str)
		fmt.Println(str)
	}
}

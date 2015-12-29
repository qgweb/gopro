package server

import (
	"errors"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/qgweb/gopro/qianzhao-ht-tcpserver/logger"
	"github.com/qgweb/gopro/qianzhao-ht-tcpserver/protocol"
)

const (
	PROTOCOL_HEAD = "qgbrower"
)

var (
	connManager    = NewAccountConnManager()
	accountManager = NewAccountManager(logger.New("qianzhaotcp"))
)

// 封包
func ProtocolPack(data []byte) []byte {
	p := protocol.NewProtocol(PROTOCOL_HEAD)
	return p.Packet(data)
}

// 解包
func ProtocolUnPack(data []byte) []byte {
	p := protocol.NewProtocol(PROTOCOL_HEAD)
	b, _ := p.Unpack(data)
	if len(b) == 0 {
		return []byte("")
	}

	return b[0]
}

type Server struct {
	cm     *ConnectionManager
	socket *net.TCPListener
	logger *logger.Logger
}

func New(logger *logger.Logger, host string, port string) (*Server, error) {
	s := &Server{cm: NewConnectionManager(), logger: logger}

	addr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%s", host, port))
	if err != nil {
		return nil, fmt.Errorf("fail to resolve addr: %v", err)
	}
	sock, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("fail to listen tcp: %v", err)
	}

	s.socket = sock
	return s, nil
}

func NewFromFD(logger *logger.Logger, fd uintptr) (*Server, error) {
	s := &Server{cm: NewConnectionManager(), logger: logger}

	file := os.NewFile(fd, "/tmp/sock-go-graceful-restart")
	listener, err := net.FileListener(file)
	if err != nil {
		return nil, errors.New("File to recover socket from file descriptor: " + err.Error())
	}
	listenerTCP, ok := listener.(*net.TCPListener)
	if !ok {
		return nil, fmt.Errorf("File descriptor %d is not a valid TCP socket", fd)
	}
	s.socket = listenerTCP

	return s, nil
}

func (s *Server) GetAccountManager() *AccountManager {
	return accountManager
}

func (s *Server) GetAccountConnManager() *AccountConnManager {
	return connManager
}

func (s *Server) Stop() {
	// Accept will instantly return a timeout error
	s.socket.SetDeadline(time.Now())
}

func (s *Server) ListenerFD() (uintptr, error) {
	file, err := s.socket.File()
	if err != nil {
		return 0, err
	}
	return file.Fd(), nil
}

func (s *Server) Wait() {
	s.cm.Wait()
}

var WaitTimeoutError = errors.New("timeout")

func (s *Server) WaitWithTimeout(duration time.Duration) error {
	timeout := time.NewTimer(duration)
	wait := make(chan struct{})
	go func() {
		s.Wait()
		wait <- struct{}{}
	}()

	select {
	case <-timeout.C:
		return WaitTimeoutError
	case <-wait:
		return nil
	}
}

func (s *Server) StartAcceptLoop() {
	for {
		conn, err := s.socket.Accept()
		if err != nil {
			if nerr, ok := err.(net.Error); ok && nerr.Timeout() {
				s.logger.Println("Stop accepting connections")
				return
			}
			s.logger.Println("[Error] fail to accept:", err)
		}
		go func() {
			s.cm.Add(1)
			s.handleConn(conn.(*net.TCPConn))
			s.cm.Done()
		}()
	}
}

func (s *Server) handleConn(conn *net.TCPConn) {
	buffer := make([]byte, 1024)
	for {
		//接收服务
		n, err := conn.Read(buffer)
		if err != nil {
			s.logger.Println("客户端连接失败，错误信息：", err)
			// 处理断开服务
			break
		}

		//s.logger.Println(buffer[0:n])

		r, err := UmRequest(ProtocolUnPack(buffer[0:n]))
		if err != nil {
			s.logger.Println("请求参数解析错误，信息为：", err)
			//conn.Write(buffer[0:n])
			continue
		}

		s.logger.Println(r)

		switch r.Action {
		case "ping": //响应ping包
			(&Event{}).RepPing(conn)
			break
		case "plink": // 连接请求(前奏)
			(&Event{}).Start(conn, &r)
			break
		case "link": // 连接
			(&Event{}).Link(conn, &r)
		case "stop": // 停止请求
			(&Event{}).Stop(conn, &r)
			break
		case "info": //监听内部程序状态
			(&Event{}).Info(conn)
		case "havebind":
			(&Event{}).HaveCardByPhone(conn, &r)
		}
	}
}

func (s *Server) Addr() net.Addr {
	return s.socket.Addr()
}

func (s *Server) ConnectionsCounter() int {
	return s.cm.Counter
}

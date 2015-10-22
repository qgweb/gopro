package main

import (
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/qgweb/gopro/qianzhao-dz-tcpserver/logger"
	"github.com/qgweb/gopro/qianzhao-dz-tcpserver/server"
)

func main() {
	var (
		log      = logger.New("qianzhao-dz-tcpserver")
		fileName = fmt.Sprintf("./data/%d.dat", os.Getpid())
		s        *server.Server
		err      error
	)

	runtime.GOMAXPROCS(runtime.NumCPU())

	if os.Getenv("_GRACEFUL_RESTART") == "true" {
		s, err = server.NewFromFD(log, 3)
	} else {
		s, err = server.New(log, "", "9092")
		d := server.GetLastFile("./data")
		if d != nil {
			server.DealData(d)
		}
	}

	if err != nil {
		log.Fatalln("fail to init server:", err)
	}

	log.Println("Listen on", s.Addr())

	go s.StartAcceptLoop()
	go s.GetAccountManager().TimeFlushDisk(fileName)
	go s.GetAccountManager().TimeCheckAccountUTime(func(name string) {
		(&server.Event{}).InnerStop(name)
	})
	go s.GetAccountConnManager().Ping(func(name string) {
		(&server.Event{}).InnerStop(name)
	})

	signals := make(chan os.Signal)
	signal.Notify(signals, syscall.SIGHUP, syscall.SIGTERM)
	for sig := range signals {
		if sig == syscall.SIGTERM {
			// Stop accepting new connections
			s.Stop()
			// 发送信号，flush到磁盘
			s.GetAccountManager().StopChan <- true
			s.GetAccountManager().StopDiskChan <- true
			<-s.GetAccountManager().OverChan
			<-s.GetAccountManager().OverChan

			// Wait a maximum of 10 seconds for existing connections to finish
			err := s.WaitWithTimeout(10 * time.Second)
			if err == server.WaitTimeoutError {
				log.Printf("Timeout when stop  ping server, %d active connections will be cut.\n", s.ConnectionsCounter())
				os.Exit(-127)
			}
			// Then the program exists
			log.Println("Server shutdown successful")
			os.Exit(0)
		} else if sig == syscall.SIGHUP {
			// Stop accepting requests
			s.Stop()
			// Get socket file descriptor to pass it to fork
			listenerFD, err := s.ListenerFD()
			if err != nil {
				log.Fatalln("Fail to get socket file descriptor:", err)
			}
			// Set a flag for the new process start process
			os.Setenv("_GRACEFUL_RESTART", "true")
			execSpec := &syscall.ProcAttr{
				Env:   os.Environ(),
				Files: []uintptr{os.Stdin.Fd(), os.Stdout.Fd(), os.Stderr.Fd(), listenerFD},
			}
			// Fork exec the new version of your server
			fork, err := syscall.ForkExec(os.Args[0], os.Args, execSpec)
			if err != nil {
				log.Fatalln("Fail to fork", err)
			}
			log.Println("SIGHUP received: fork-exec to", fork)
			// Wait for all conections to be finished
			s.Wait()
			log.Println(os.Getpid(), "Server gracefully shutdown")

			// Stop the old server, all the connections have been closed and the new one is running
			os.Exit(0)
		}
	}
}

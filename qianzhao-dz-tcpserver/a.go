package main

import (
	"log"
	"time"

	"github.com/qgweb/gopro/qianzhao-dz-tcpserver/logger"
	"github.com/qgweb/gopro/qianzhao-dz-tcpserver/server"
)

func show(list map[string]*server.Account) {
	for k, v := range list {
		log.Println("===============")
		log.Println("name:", k)
		log.Println("begtin_time", time.Unix(v.BTime, 0).Format("2006-01-02 15:04:05"))
		log.Println("end_time", time.Unix(v.ETime, 0).Format("2006-01-02 15:04:05"))
		log.Println("can_time", v.CTime)
		log.Println("===============")
	}
}

func main1() {
	log := logger.New("Server")
	am := server.NewAccountManager(log)
	u := &server.Account{}
	u.Name = "zb1"
	u.BTime = time.Now().Unix()
	u.ETime = 0
	u.CTime = 30
	am.Add(u)
	u1 := &server.Account{}
	u1.Name = "zb2"
	u1.BTime = time.Now().Add(time.Minute).Unix()
	u1.ETime = 0
	u1.CTime = 30
	am.Add(u1)

	show(am.Range())

	go am.TimeFlushDisk("aaaa.dat")
	go am.TimeCheckAccountUTime(func(u string) {
		log.Println(u)
	})

	go func() {
		<-time.After(time.Minute)
		close(am.StopChan)
	}()
	select {}
}

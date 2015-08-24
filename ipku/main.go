package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"runtime"
	"strings"

	"github.com/wangtuanjie/ip17mon"
	"gopkg.in/ini.v1"
)

var (
	area    = make(map[string]string)
	err     error
	iniFile *ini.File
)

func init() {
	confData, err := ioutil.ReadFile("./conf.ini")
	if err != nil {
		log.Fatalln("配置文件不存在,", err)
	}

	iniFile, _ = ini.Load(confData)
	ipFile := iniFile.Section("file").Key("ip").String()
	data, err := ioutil.ReadFile(ipFile)
	if err != nil {
		log.Fatalln("读取ip库文件失败,", err)
	}

	ip17mon.InitWithData(data)
	initArea()

}

func initArea() {
	areaFile := iniFile.Section("file").Key("area").String()
	f, _ := os.Open(areaFile)
	defer f.Close()

	bi := bufio.NewReader(f)
	for {
		l, err := bi.ReadString('\n')
		if err == io.EOF {
			break
		}

		ls := strings.Split(strings.TrimSpace(l), "\t")
		area[ls[0]] = ls[1]
	}
}

func getIp(w http.ResponseWriter, r *http.Request) {
	ip := r.URL.Query().Get("sspip")
	if ip == "" {
		w.Write([]byte(""))
		return
	}
	loc, err := ip17mon.Find(ip)
	if err != nil {
		log.Println(err)
		w.Write([]byte(""))
		return
	}

	region := ""
	city := ""
	if v, ok := area[loc.Region]; ok {
		region = v
	}

	if v, ok := area[loc.City]; ok {
		city = v
	} else {
		for k, v := range area {
			if strings.Contains(loc.City, k) {
				city = v
				break
			}
		}
	}

	w.Write([]byte(region + "_" + city))
}

func main() {
	var (
		host = iniFile.Section("http").Key("host").String()
		port = iniFile.Section("http").Key("port").String()
	)

	runtime.GOMAXPROCS(runtime.NumCPU())

	http.HandleFunc("/getip", getIp)
	http.ListenAndServe(fmt.Sprintf("%s:%s", host, port), nil)
}

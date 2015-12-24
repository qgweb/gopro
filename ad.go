package main

import (
	"bufio"
	"flag"
	"github.com/ngaut/log"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"io"
	"os"
	"strings"
)

var (
	conf = flag.String("conf", "", "文件")
)

func init() {
	flag.Parse()
}

func main() {
	db, err := mgo.Dial("192.168.0.93:10001/user_cookie")
	if err != nil {
		log.Fatal(err)
		return
	}
	defer db.Close()

	var (
		coxNum      int64
		coxTotalNum int64
		num         int64
	)

	f, err := os.Open(*conf)
	if err != nil {
		log.Fatal(err)
		return
	}
	defer f.Close()

	bf := bufio.NewReader(f)
	for {
		num++
		line, err := bf.ReadString('\n')
		if err == io.EOF {
			break
		}
		cox := strings.TrimSpace(line)

		n, _ := db.DB("user_cookie").C("dt_user").Find(bson.M{"pid": "31", "cox": cox}).Count()
		if n > 0 {
			coxTotalNum = coxTotalNum + int64(n)
			coxNum = coxNum + 1
		}

		if num%10000 == 0 {
			log.Error(num, coxNum, coxTotalNum)
		}
	}
	log.Error(num, coxNum, coxTotalNum)
}

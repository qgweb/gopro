package main

import (
	"fmt"
	"log"
	"time"

	"github.com/qgweb/gopro/lib/convert"
	"github.com/tealeg/xlsx"
	"gopkg.in/mgo.v2"
)

func main() {
	t := time.Now()
	h := "23"
	fmt.Println(time.Date(t.Year(), t.Month(), t.Day(), convert.ToInt(h), 0, 0, 0, time.Local).Unix())

	return

	sess, err := mgo.Dial("192.168.1.199:27017/data_source")
	if err != nil {
		log.Fatalln(err)
	}

	f := xlsx.NewFile()
	s, _ := f.AddSheet("sheet1")

	it := sess.DB("data_source").C("domain_category").Find(nil).Iter()
	var data map[string]interface{}
	for it.Next(&data) {
		r := s.AddRow()
		r.AddCell().SetValue(data["name"].(string))
		r.AddCell().SetValue(data["cid"].(string))
		log.Println(data["name"], data["cid"])
	}
	it.Close()
	f.Save("./cat.xls")
}

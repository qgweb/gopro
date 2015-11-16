package main

import (
	"github.com/tealeg/xlsx"
	"gopkg.in/mgo.v2"
	"log"
)

func main() {

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

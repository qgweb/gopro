package main

import (
	"gopkg.in/mgo.v2"
	"github.com/ngaut/log"
	"time"
	"gopkg.in/mgo.v2/bson"
	"fmt"
	"github.com/qgweb/gopro/lib/convert"
)

func main() {
	//1451232000
	//192.168.0.93:10003
	sess, err := mgo.Dial("192.168.0.93:10003/user_cookie")
	if err !=nil {
		log.Fatal(err)
		return
	}
	defer sess.Close()

	sess.SetCursorTimeout(0)
	sess.SetSocketTimeout(time.Hour)
	sess.SetSyncTimeout(time.Hour)

	var count = 0
	iter:=sess.DB("user_cookie").C("dt_user").Find(bson.M{}).Select(bson.M{"_id":1,"date":1}).Iter()
	for {
		count++
		if count % 10000 == 0{
			fmt.Println("###################")
		}
		var info map[string]interface{}
		if !iter.Next(&info) {
			break
		}
		if convert.ToInt64(info["date"]) >= int64(1451232000) {
			fmt.Println(info["_id"].(bson.ObjectId).Hex(),"|",info["date"])
		}
	}
}

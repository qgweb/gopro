package main

import (
	"github.com/ngaut/log"
	"time"
)

func main() {
	a, _ := time.ParseInLocation("2006-01-02", time.Now().Add(-time.Hour*24).Format("2006-01-02"), time.Local)
	log.Error(a.Unix())
}

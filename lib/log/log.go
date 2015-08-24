package log

import (
	"log"
)

type ZLog struct {
	msg chan string
}

func NewLog(chanSize int) *ZLog {
	if chanSize < 0 {
		panic("chanSize不能小于0")
	}
	zl := &ZLog{msg: make(chan string, chanSize)}
	go zl.run()
	return zl
}

func (this *ZLog) run() {
	for {
		select {
		case m := <-this.msg:
			log.Println(m)
		}

	}
}

func (this *ZLog) writeMsg(msg string) {
	this.msg <- msg
}

func (this *ZLog) Warn(msg string) {
	this.writeMsg("[Warn] " + msg)
}

func (this *ZLog) Error(msg string) {
	this.writeMsg("[Error] " + msg)
}

func (this *ZLog) Info(msg string) {
	this.writeMsg("[Info] " + msg)
}

func (this *ZLog) Notice(msg string) {
	this.writeMsg("[Notice] " + msg)
}

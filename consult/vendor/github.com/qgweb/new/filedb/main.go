package main

import (
	"flag"
	"fmt"
	"github.com/robfig/cron"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

const (
	EXPIRETIME = 86400 * 3
)

var (
	httphost = flag.String("host", "127.0.0.1", "http地址")
	httpport = flag.String("port", "3344", "http端口")
	dbpath   = flag.String("path", "/tmp/", "数据存储地址")
	mux      = sync.Mutex{}
)

func init() {
	flag.Parse()
}

func recv(w http.ResponseWriter, r *http.Request) {
	mux.Lock()
	defer mux.Unlock()
	r.ParseMultipartForm(32 << 20)
	file, handler, err := r.FormFile("name")
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
		return
	}
	defer file.Close()

	f, err := os.OpenFile(*dbpath+"/"+handler.Filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
		return
	}
	defer f.Close()
	io.Copy(f, file)
	w.Write([]byte("ok"))
}

func get(w http.ResponseWriter, r *http.Request) {
	filename := r.URL.Query().Get("name")
	f, err := os.Open(*dbpath + "/" + filename)
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
		return
	}
	defer f.Close()

	io.Copy(w, f)
}

func clean() {
	f, err := os.Open(*dbpath)
	if err != nil {
		log.Println(err)
		return
	}
	defer f.Close()

	fs, err := f.Readdir(0)
	if err != nil {
		log.Println(err)
		return
	}

	for _, v := range fs {
		if time.Since(v.ModTime()).Seconds() > EXPIRETIME {
			fmt.Println(v.Name())
			os.Remove(*dbpath + "/" + v.Name())
		}
	}
}

func crontab() {
	c := cron.New()
	c.Start()
	c.AddFunc("* */1 * * * *", clean)
}

func main() {
	crontab()
	http.HandleFunc("/recv", recv)
	http.HandleFunc("/get", get)
	http.ListenAndServe(fmt.Sprintf("%s:%s", *httphost, *httpport), nil)
}

package main

import (
	"encoding/json"
	"flag"
	"github.com/ngaut/log"
	"html/template"
	"io/ioutil"
	"net/http"
)

var (
	hm *HostsManager
)

func init() {
	hm = NewHostsManager()
}

func AddRecord(w http.ResponseWriter, r *http.Request) {
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Error(err)
		return
	}
	h := Hosts{}

	err = json.Unmarshal(data, &h)
	if err != nil {
		log.Error(err)
		return
	}

	hm.Add(&h)
}

func Range(w http.ResponseWriter, r *http.Request) {
	t := template.New("index.html")
	t, _ = t.ParseFiles("index.html")
	t.Execute(w, hm.Range())
}

func main() {
	var port = flag.String("port", "8112", "端口")
	flag.Parse()

	go hm.Run()
	http.HandleFunc("/add", AddRecord)
	http.HandleFunc("/", Range)
	log.Error(http.ListenAndServe(":"+*port, nil))
}

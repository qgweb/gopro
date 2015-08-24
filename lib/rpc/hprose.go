package rpc

import (
	"bytes"
	"net/http"

	"goclass/encrypt"

	"github.com/hprose/hprose-go/hprose"
)

type MyServer struct {
	ser      *hprose.HttpService
	key      string
	httpKey  string
	httpCode string
}

func NewMyServer(ser *hprose.HttpService, key string, httpkey string, httpcode string) *MyServer {
	return &MyServer{ser, key, httpkey, httpcode}
}

func (this *MyServer) md5(str string) string {
	return encrypt.GetEncoder(encrypt.TYPE_MD5).Encode(str)
}

func (this *MyServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if this.md5(this.md5(r.Header.Get(this.httpKey))+this.key) == r.Header.Get(this.httpCode) {
		this.ser.ServeHTTP(w, r)
	} else {
		buf := new(bytes.Buffer)
		writer := hprose.NewWriter(buf, true)
		writer.Stream().WriteByte(hprose.TagError)
		writer.WriteString("无权限")
		writer.Stream().WriteByte(hprose.TagEnd)
		w.Write(buf.Bytes())
	}
}

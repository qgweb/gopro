//webp格式转换
// 需要webpconv程序
package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"time"
)

var (
	host = flag.String("host", "127.0.0.1", "host url")
	port = flag.String("port", "8888", "port")
)

func init() {
	flag.Parse()
}
func main() {
	http.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
		file, h, err := r.FormFile("file")
		if err != nil {
			log.Fatal("FormFile: ", err.Error())
			return
		}
		defer func() {
			if err := file.Close(); err != nil {
				log.Fatal("Close: ", err.Error())
				return
			}
		}()

		body, err := ioutil.ReadAll(file)
		if err != nil {
			log.Fatal("ReadAll: ", err.Error())
			return
		}

		//处理图片
		rd := rand.New(rand.NewSource(time.Now().UnixNano()))
		randNum := rd.Intn(999999999)
		fileName := fmt.Sprintf("/tmp/%s_%d.jpg", h.Filename, randNum)
		outFileName := fmt.Sprintf("/tmp/%s_%d.webp", h.Filename, randNum)
		err = ioutil.WriteFile(fileName, body, os.ModePerm)
		if err != nil {
			log.Println(err)
			w.Write([]byte(""))
			return
		}

		cmd := exec.Command("webpconv", "-quality", "85", fileName)
		cmd.Run()
		os.Remove(fileName)
		data, err := ioutil.ReadFile(outFileName)
		if err != nil {
			log.Println(err)
			w.Write([]byte(""))
			return
		}
		os.Remove(outFileName)
		w.Write(data)
	})

	log.Fatalln(http.ListenAndServe(fmt.Sprintf("%s:%s", *host, *port), nil))
}

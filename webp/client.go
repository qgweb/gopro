//webp格式
//请求服务端 转换图片然后保存到本地
package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
)

//read file data
func readFile(name string) ([]byte, error) {
	d, err := ioutil.ReadFile(name)
	if err != nil {
		return nil, err
	}

	return d, nil
}

func newfileUploadRequest(uri string, params map[string]string, paramName, path string) (*http.Request, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile(paramName, filepath.Base(path))
	if err != nil {
		return nil, err
	}
	_, err = io.Copy(part, file)

	for key, val := range params {
		_ = writer.WriteField(key, val)
	}
	err = writer.Close()
	if err != nil {
		return nil, err
	}

	r, err := http.NewRequest("POST", uri, body)
	r.Header.Set("Content-Type", writer.FormDataContentType())
	return r, err
}

//post data to server
func postFile(filePath string, host string) (string, error) {
	extraParams := make(map[string]string)
	request, err := newfileUploadRequest(host, extraParams, "file", filePath)
	if err != nil {
		log.Fatal(err)
	}
	client := &http.Client{}

	resp, err := client.Do(request)
	if err != nil {
		log.Fatal(err)
	} else {
		body := &bytes.Buffer{}
		_, err := body.ReadFrom(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		resp.Body.Close()

		if len(body.Bytes()) != 0 {
			newfileName := filePath + ".webp"
			ioutil.WriteFile(newfileName, body.Bytes(), os.ModePerm)
		}
	}
	return "", nil
}

var (
	inputFile = flag.String("infile", "", "input file")
	host      = flag.String("host", "", "host url")
	port      = flag.String("port", "", "port")
)

func init() {
	flag.Parse()
}
func main() {
	if *inputFile == "" {
		log.Fatalln("文件不能为空")
	}

	runtime.GOMAXPROCS(runtime.NumCPU())
	postFile(*inputFile, fmt.Sprintf("http://%s:%s/upload", *host, *port))
}

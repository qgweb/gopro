package httpsqs

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	nurl "net/url"
)

type Httpsqs struct {
	host    string
	port    string
	auth    string
	queue   string
	method  string //http请求方式 默认为 "GET"
	charset string //http头字符编码 默认utf-8  常见的有 utf-8、gb2312、gbk、gb18030、big5
	opt     string //最后一次操作
	url     string //最后一次请求的完整url
}

const (
	UNKNOW_ERROR  = "Unknown error"                  //未知错误
	NO_METHOD     = "No this method"                 //没有这个方法
	PARAMS_EMPTY  = "Host,Port,Auth cannot be empty" //新建实例参数不能为空
	NO_QUEUE_NAME = "Please set queue name"          //没有设置队列名

	HTTPSQS_PUT_OK    = "HTTPSQS_PUT_OK"    //入队列成功
	HTTPSQS_PUT_ERROR = "HTTPSQS_PUT_ERROR" //入队列失败
	HTTPSQS_PUT_END   = "HTTPSQS_PUT_END"   //队列已满

	HTTPSQS_GET_END = "HTTPSQS_GET_END" //未取出数据

	HTTPSQS_RESET_OK    = "HTTPSQS_RESET_OK"    //重置队列成功
	HTTPSQS_RESET_ERROR = "HTTPSQS_RESET_ERROR" //重置失败

	HTTPSQS_MAXQUEUE_OK     = "HTTPSQS_MAXQUEUE_OK"     //更改最大队列数量成功
	HTTPSQS_MAXQUEUE_CANCEL = "HTTPSQS_MAXQUEUE_CANCEL" //更改最大队列数量取消

	HTTPSQS_SYNCTIME_OK     = "HTTPSQS_SYNCTIME_OK"     //修改定时刷新内存缓冲区内容到磁盘的间隔时间 成功
	HTTPSQS_SYNCTIME_CANCEL = "HTTPSQS_SYNCTIME_CANCEL" //操作取消

	HTTPSQS_AUTH_FAILED = "HTTPSQS_AUTH_FAILED" //密码校验失败
	HTTPSQS_ERROR       = "HTTPSQS_ERROR"       //发生全局错误
)

func New(host, port, auth string) (*Httpsqs, error) {
	if host == "" || port == "" || auth == "" {
		return nil, errors.New(PARAMS_EMPTY)
	}
	return &Httpsqs{host, port, auth, "", "GET", "utf-8", "", ""}, nil
}

/**
 * 入队列
 */
func (hs *Httpsqs) Put(param map[string]string) error {
	if hs.queue == "" {
		return errors.New(NO_QUEUE_NAME)
	}

	if hs.method == "GET" {
		hs.opt = "put"

		data, _ := json.Marshal(param)
		url := fmt.Sprintf("http://%s:%s/?name=%s&opt=%s&data=%s&auth=%s", hs.host, hs.port, hs.queue, hs.opt, nurl.QueryEscape(string(data)), hs.auth)
		hs.url = url
		resp, err := http.Get(url)
		defer resp.Body.Close()
		if err != nil {
			return err
		}
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		switch string(body) {
		case HTTPSQS_PUT_OK:
			err = nil
		case HTTPSQS_PUT_ERROR:
			err = errors.New(HTTPSQS_PUT_ERROR)
		case HTTPSQS_PUT_END:
			err = errors.New(HTTPSQS_PUT_END)
		default:
			err = errors.New(UNKNOW_ERROR)
		}
		return err
	} else if hs.method == "POST" {

	} else {
		return errors.New(NO_METHOD)
	}

	return nil
}

/**
 * 出队列
 */
func (hs *Httpsqs) Get() {}

func (hs *Httpsqs) SetQueue(name string) error {
	if name == "" {
		return errors.New("queue name cannot be empty")
	}
	hs.queue = name
	return nil
}

func (hs *Httpsqs) GetLastUrl() string {
	return hs.url
}

func (hs *Httpsqs) GetLastOpt() string {
	return hs.opt
}

package mulgrab

import (
	"crypto/tls"
	"github.com/ngaut/log"
	"github.com/qgweb/gopro/lib/mulgrab/agent"
	"golang.org/x/net/html/charset"
	"io"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"sync"
	"time"
)

type MulGrab struct {
	client *http.Client
	mux    sync.Mutex
}

type Config struct {
	DialTimeout     time.Duration
	DeadlineTimeOut time.Duration
}

func New(conf Config) *MulGrab {
	client := &http.Client{
		CheckRedirect: func(req *http.Request, _ []*http.Request) error {
			return nil
		},
	}

	transport := &http.Transport{
		Dial: func(network, addr string) (net.Conn, error) {
			c, err := net.DialTimeout(network, addr, conf.DialTimeout)
			if err != nil {
				log.Error(err)
				return nil, err
			}
			c.SetDeadline(time.Now().Add(conf.DeadlineTimeOut))
			return c, nil
		},
	}

	transport.TLSClientConfig = &tls.Config{RootCAs: nil, InsecureSkipVerify: true}
	transport.DisableCompression = true
	client.Transport = transport

	return &MulGrab{client, sync.Mutex{}}
}

func (this *MulGrab) changeCharsetEncodingAuto(sor io.ReadCloser, contentTypeStr string) (string, error) {
	var err error
	destReader, err := charset.NewReader(sor, contentTypeStr)

	if err != nil {
		destReader = sor
	}

	var sorbody []byte
	if sorbody, err = ioutil.ReadAll(destReader); err != nil {
		return "", err
	}
	return string(sorbody), nil
}

func (this *MulGrab) Get(url string, head http.Header) (string, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	if head == nil {
		req.Header = this.GetHeader()
	} else {
		req.Header = head
	}

	this.mux.Lock()
	resp, err := this.client.Do(req)
	this.mux.Unlock()
	if err != nil {
		return "", err
	}

	if resp.Body != nil {
		defer resp.Body.Close()
		return this.changeCharsetEncodingAuto(resp.Body, resp.Header.Get("Content-Type"))
	}
	return "", nil
}

func (this *MulGrab) GetHeader() (head http.Header) {
	head = make(http.Header)
	l := len(agent.UserAgents["common"])
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	head.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	head.Set("Accept-Language", "zh-CN,zh;q=0.8")
	head.Set("Cache-Control", "no-cache")
	//head.Set("Connection", "keep-alive:300")
	head.Set("Host", "www.taobao.com")
	head.Set("Pragma", "no-cache")
	head.Set("User-Agent", agent.UserAgents["common"][r.Intn(l)])
	return
}

package http

import (
	"github.com/qgweb/new/lib/encrypt"
	"github.com/qgweb/new/lib/timestamp"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"github.com/juju/errors"
)

func Post(host string, param map[string]string, key string) ([]byte, error) {
	uv := url.Values{}

	for k, v := range param {
		uv.Set(k, v)
	}

	r, err := http.NewRequest("POST", host, ioutil.NopCloser(strings.NewReader(uv.Encode())))
	if err != nil {
		return nil, err
	}

	var t = timestamp.GetTimestamp()
	r.Header.Add("8706C971C8BE6CDE39DDEF6A39C36572", t)
	r.Header.Add("6B7D574EAA7A0F332F05624E17F4BE01", encrypt.DefaultMd5.Encode(t+key))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded;")

	req, err := http.DefaultClient.Do(r)
	if err != nil {
		return nil, err
	}

	if req.Body != nil {
		defer req.Body.Close()
		return ioutil.ReadAll(req.Body)
	}

	return nil , errors.New("无数据")
}

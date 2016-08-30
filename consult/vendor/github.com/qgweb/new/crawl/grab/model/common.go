package model

import (
	"net/url"
)

type Crawler interface {
	Grab(string) map[string]interface{}
}

type DataStreamer interface {
	Receive()
	Dispatch(url.Values)
	Save([]byte)
}


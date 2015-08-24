package encrypt

import "net/url"

type UrlCode struct {
}

func (this UrlCode) Encode(source string) string {
	return url.QueryEscape(source)
}

func (this UrlCode) Decode(source string) string {
	res, _ := url.QueryUnescape(source)
	return res
}

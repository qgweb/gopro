package encrypt

import "encoding/base64"

type Base64 struct {
}

func (this Base64) Encode(source string) string {
	return base64.StdEncoding.EncodeToString([]byte(source))
}

func (this Base64) Decode(source string) string {
	res, _ := base64.StdEncoding.DecodeString(source)
	return string(res)
}

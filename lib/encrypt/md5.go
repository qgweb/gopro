package encrypt

import (
	"crypto/md5"
	"encoding/hex"
)

type Md5 struct {
}

func (this Md5) Encode(source string) string {
	m5 := md5.New()
	m5.Write([]byte(source))
	return hex.EncodeToString(m5.Sum([]byte("")))
}

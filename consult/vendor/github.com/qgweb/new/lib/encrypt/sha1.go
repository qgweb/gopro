package encrypt

import (
	"crypto/sha1"
	"encoding/hex"
)

type SHA1 struct {
}

func (this SHA1) Encode(source string) string {
	m5 := sha1.New()
	m5.Write([]byte(source))
	return hex.EncodeToString(m5.Sum([]byte("")))
}

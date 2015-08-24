package encrypt

const (
	TYPE_MD5    = "md5"
	TYPE_BASE64 = "base64"
	TYPE_URL    = "url"
	TYPE_AES    = "aes"
)

var (
	DefaultBase64  Base64
	DefaultMd5     Md5
	DefaultUrlcode UrlCode
	DefaultAes     Aes
)

//不可逆加密
type Encoder interface {
	Encode(source string) string
}

//可逆加密
type EnDeCoder interface {
	Encode(source string) string
	Decode(source string) string
}

//可逆加密 有私钥
type EnDeKeyCoder interface {
	Encode(source, key []byte) ([]byte, error)
	Decode(source, key []byte) ([]byte, error)
}

//获取不可逆加密接口
func GetEncoder(t string) Encoder {
	switch t {
	case TYPE_MD5:
		return Md5{}
	}
	return nil
}

//获取可逆加密接口
func GetEnDecoder(t string) EnDeCoder {
	switch t {
	case TYPE_BASE64:
		return Base64{}
	case TYPE_URL:
		return UrlCode{}
	}
	return nil
}

//获取可逆加密接口
func GetEnDeKeycoder(t string) EnDeKeyCoder {
	switch t {
	case TYPE_AES:
		return Aes{}
	}
	return nil
}

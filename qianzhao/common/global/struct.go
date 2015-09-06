// 定义接口的结构
package global

// BEGIN ====== 用户账户查询结构 =================
type getUserProductResponse struct {
	Result               string `xml:"result"`
	GetUserProductReturn string `xml:"getUserProductReturn"`
}

type UserEnvelope struct {
	By UserBody `xml:"Body"`
}

type UserBody struct {
	GP getUserProductResponse `xml:"getUserProductResponse"`
}

// END ====== 用户账户查询结构 =================

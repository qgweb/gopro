// 定义接口的结构
package global

// BEGIN ====== 用户账户查询结构 =================
type getUserProductResponse struct {
	Result               string `xml:"result"`
	GetUserProductReturn string `xml:"getUserProductReturn"`
}

type Envelope struct {
	By Body `xml:"Body"`
}

type Body struct {
	GP getUserProductResponse `xml:"getUserProductResponse"`
}

// END ====== 用户账户查询结构 =================

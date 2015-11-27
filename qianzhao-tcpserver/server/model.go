// 具体业务罗辑
package server

import (
	"bytes"
	"fmt"
	"github.com/astaxie/beego/httplib"
	"github.com/qgweb/gopro/lib/convert"
	"github.com/qgweb/gopro/qianzhao-tcpserver/model"
	"log"
	"strings"
	"text/template"
	"time"
)

const (
	USER_QUERY_URL       = "http://js.vnet.cn/ProvinceForSPSearchUserName/services/ProvinceForSPSearchUserName?wsdl"
	USER_QUERY_PRODUCTID = "1100099900000000"
	DEFAULT_SPEED        = 1024 * 30 * 1.0 // 单位kb
	QOS_QUERY_URL        = "http://202.102.41.31:8080/services/DacsForSPInterface"
)

var (
	ErrProgram         = &ErrBrand{"5000", "程序发生异常", ""}
	ErrNotJiangShuUser = &ErrBrand{"5001", "须为江苏电信校园宽带用户", ""}
	ErrNotAllowUP      = &ErrBrand{"5002", "当前宽带环境不满足提速条件", ""}
	ErrTimeOut         = &ErrBrand{"5003", "体验时间已结束", ""}
	ErrUpFaile         = &ErrBrand{"5004", "提速失败", ""}
	ErrStopFaile       = &ErrBrand{"5004", "停止失败", ""}
	/////
	AreaCodes = map[string]string{
		"南京":  "0001",
		"苏州":  "0002",
		"无锡":  "0003",
		"常州":  "0004",
		"镇江":  "0005",
		"扬州":  "0006",
		"南通":  "0007",
		"泰州":  "0008",
		"徐州":  "0009",
		"淮安":  "0010",
		"盐城":  "0011",
		"连云港": "0012",
		"宿迁":  "0013",
	}
)

type ErrBrand struct {
	Code    string
	Msg     string
	Content string
}

type UserData struct {
	Area    string
	Account string
}

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

/**
 * <x:Envelope xmlns:x="http://schemas.xmlsoap.org/soap/envelope/" xmlns:sp="http://sp.interfaces.pccs.linkage.com" xmlns:req="http://req.sp.interfaces.pccs.linkage.com">
  <x:Header/>
  <x:Body>
    <sp:startQos>
      <sp:req>
        <req:accNbr>0</req:accNbr>
        <req:area>001</req:area>
        <req:policycode>101803030</req:policycode>
        <req:serverip>122.225.98.80</req:serverip>
        <req:spcode>haobai</req:spcode>
        <req:sppassword>haobai123</req:sppassword>
        <req:times>0</req:times>
        <req:type>0</req:type>
        <req:spip>192.15.15.2</req:spip>
        <req:userInfos>
          <sp:item>
            <req:userip>180.98.1.2</req:userip>
            <req:userport>33</req:userport>
            <req:username>dacs_test02</req:username>
          </sp:item>
        </req:userInfos>
      </sp:req>
    </sp:startQos>
  </x:Body>
</x:Envelope>
*/
type RequestQos struct {
	AccNbr     string
	Area       string
	Policycode string
	Serverip   string
	Spcode     string
	Sppassword string
	Times      string
	Type       string
	Spip       string
	UserInfos  UserInfo
}

type UserInfo struct {
	Userip   string `xml:"userip"`
	Userport string `xml:"userport"`
	Username string `xml:"username"`
}

/**
 * <?xml version="1.0" encoding="UTF-8"?>
<soapenv:Envelope xmlns:soapenv="http://schemas.xmlsoap.org/soap/envelope/" xmlns:xsd="http://www.w3.org/2001/XMLSchema" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
  <soapenv:Body>
    <startQosResponse xmlns="http://sp.interfaces.pccs.linkage.com">
      <startQosReturn>
        <ns1:userResponses xmlns:ns1="http://rsp.sp.interfaces.pccs.linkage.com">
          <item>
            <ns1:errordescription>&#x7528;&#x6237;&#x4E0D;&#x5728;&#x7EBF;</ns1:errordescription>
            <ns1:result>-10012</ns1:result>
            <ns1:userip>180.98.1.2</ns1:userip>
            <ns1:username>dacs_test02</ns1:username>
          </item>
        </ns1:userResponses>
      </startQosReturn>
    </startQosResponse>
  </soapenv:Body>
</soapenv:Envelope>
*/

// 开启
type QosStartEnvelope struct {
	By QosStartBody `xml:"Body"`
}

type QosStartBody struct {
	GP StartQosResponse `xml:"startQosResponse"`
}

type StartQosResponse struct {
	Rep StartQosReturn `xml:"startQosReturn"`
}

type StartQosReturn struct {
	URP UserResponses `xml:"userResponses"`
}

type UserResponses struct {
	Item ItemuserResponses `xml:"item"`
}

type ItemuserResponses struct {
	Errordescription string `xml:"errordescription"`
	Result           string `xml:"result"`
	Userip           string `xml:"userip"`
	Username         string `xml:"username"`
}

// 关闭
type QosStopEnvelope struct {
	By QosStopBody `xml:"Body"`
}

type QosStopBody struct {
	GP StopQosResponse `xml:"stopQosResponse"`
}

type StopQosResponse struct {
	Rep StopQosReturn `xml:"stopQosReturn"`
}

type StopQosReturn struct {
	URP UserResponses `xml:"userResponses"`
}

// 宽带接口管理
type BDInterfaceManager struct{}

func (this *BDInterfaceManager) CanStart(account string, addr string) Respond {
	err := this.canOpen(account, addr)
	switch err {
	case ErrProgram:
		return Respond{"500", "程序发生异常"}
	case ErrNotAllowUP, ErrNotJiangShuUser:
		return Respond{"502", "非体验用户"}
	case ErrTimeOut:
		return Respond{"501", "用户体验已结束"}
	}

	userData := strings.Split(err.Content, "|")
	userCanTime := userData[0]
	userArea := "0001"
	if len(userData) > 1 {
		userArea = userData[1]
	}

	addrAry := strings.Split(addr, ":")
	reqQos := RequestQos{}
	reqQos.AccNbr = fmt.Sprintf("%d", time.Now().Unix())
	reqQos.Area = userArea
	reqQos.Policycode = "101803030"
	reqQos.Serverip = "202.102.41.115"
	reqQos.Spcode = "qianzhao"
	reqQos.Sppassword = "qianzhao"
	reqQos.Times = "0"
	reqQos.Type = "0"
	reqQos.Spip = "180.98.1.2"
	reqQos.UserInfos.Userip = addrAry[0]
	reqQos.UserInfos.Username = account
	reqQos.UserInfos.Userport = addrAry[1]
	err = this.QosQuery(reqQos, "startQos")

	switch err {
	case ErrStopFaile:
		return Respond{"502", "非体验用户"}
	}

	return Respond{"200", userCanTime + "|" + userArea}
}

// 判断是否满足开启条件
func (this *BDInterfaceManager) canOpen(account string, addr string) *ErrBrand {
	udata, err := this.userQuery(addr)
	if err != nil {
		return err
	}
	//udata := UserData{"", "10327158472"}

	// 检测用户是否在白名单内
	baModel := model.BrandAccount{}
	if !baModel.AccountExist(model.BrandAccount{Account: udata.Account, Area: udata.Area}) {
		return ErrNotAllowUP
	}

	// 检测用户时长
	ba, err1 := baModel.GetAccountInfo(udata.Account)
	if err1 != nil || ba.Id == "" {
		return ErrProgram
	}

	// 可以使用时长
	canTime := ba.TotalTime - ba.UsedTime
	if canTime == 0 {
		return ErrTimeOut
	}

	return &ErrBrand{"200", "ok", convert.ToString(canTime) + "|" + udata.Area}
}

// 用户宽带查询(username string, areacode string)
func (this *BDInterfaceManager) userQuery(addr string) (*UserData, *ErrBrand) {
	req := httplib.Post(USER_QUERY_URL)
	req.Header("Content-Type", "text/xml; charset=utf-8")
	req.Header("SOAPAction", USER_QUERY_URL)
	req.Body(createUserSOAPXml(USER_QUERY_PRODUCTID, addr))
	a := UserEnvelope{}
	err := req.ToXml(&a)
	if err != nil {
		log.Println("[BroadBand userQuery] 解析xml失败 ", err)
		return &UserData{}, ErrProgram
	}

	res := a.By.GP.GetUserProductReturn
	resAry := strings.Split(res, "|")
	log.Println(res)
	if len(resAry) != 3 {
		return &UserData{}, ErrProgram
	}

	if resAry[0] == "0" {
		return &UserData{resAry[1], resAry[2]}, nil
	}
	if resAry[0] == "-999" {
		return &UserData{}, ErrNotJiangShuUser
	}
	if resAry[0] == "-998" {
		return &UserData{}, ErrProgram
	}

	return &UserData{}, nil
}

// 提速接口
func (this *BDInterfaceManager) QosQuery(reqQos RequestQos, method string) *ErrBrand {
	req := httplib.Post(QOS_QUERY_URL)
	req.Header("Content-Type", "text/xml; charset=utf-8")
	req.Header("SOAPAction", QOS_QUERY_URL)
	req.Body(createQosSOAPXML(reqQos, method))

	if method == "stopQos" {
		res := QosStopEnvelope{}
		req.ToXml(&res)
		log.Println(req.String())
		if strings.TrimSpace(res.By.GP.Rep.URP.Item.Result) == "0" {
			return &ErrBrand{"200", "", ""}
		} else {
			return ErrUpFaile
		}
	} else {
		res := QosStartEnvelope{}
		req.ToXml(&res)
		log.Println(req.String())
		if strings.TrimSpace(res.By.GP.Rep.URP.Item.Result) == "0" {
			return &ErrBrand{"200", "", ""}
		} else {
			return ErrUpFaile
		}
	}
}

// 停止接口
func (this *BDInterfaceManager) Stop(account string, area string, addr string) Respond {
	addrAry := strings.Split(addr, ":")

	reqQos := RequestQos{}
	reqQos.AccNbr = fmt.Sprintf("%d", time.Now().Unix())
	reqQos.Area = area
	reqQos.Policycode = "101803030"
	reqQos.Serverip = "202.102.41.115"
	reqQos.Spcode = "qianzhao"
	reqQos.Sppassword = "qianzhao"
	reqQos.Type = "0"
	reqQos.Spip = "180.98.1.2"
	reqQos.UserInfos.Userip = addrAry[0]
	reqQos.UserInfos.Username = account
	reqQos.UserInfos.Userport = addrAry[1]
	err := this.QosQuery(reqQos, "stopQos")

	switch err {
	case ErrStopFaile:
		return Respond{"500", "程序发生异常"}
	}

	return Respond{"200", "ok"}
}

// 加速请求接口
func createQosSOAPXML(data RequestQos, method string) string {
	type NewData struct {
		Req     RequestQos
		Method  string
		IsStart bool
	}

	soapBody := `<x:Envelope xmlns:x="http://schemas.xmlsoap.org/soap/envelope/" xmlns:sp="http://sp.interfaces.pccs.linkage.com" xmlns:req="http://req.sp.interfaces.pccs.linkage.com">
  <x:Header/>
  <x:Body>
    <sp:{{.Method}}>
      <sp:req>
        <req:accNbr>{{.Req.AccNbr}}</req:accNbr>
        <req:area>{{.Req.Area}}</req:area>
        <req:policycode>{{.Req.Policycode}}</req:policycode>
        <req:serverip>{{.Req.Serverip}}</req:serverip>
        <req:spcode>{{.Req.Spcode}}</req:spcode>
        <req:sppassword>{{.Req.Sppassword}}</req:sppassword>
        {{if .IsStart}}
        <req:times>{{.Req.Times}}</req:times>
        {{end}}
        <req:type>{{.Req.Type}}</req:type>
        <req:spip>{{.Req.Spip}}</req:spip>
        <req:userInfos>
          <sp:item>
            <req:userip>{{.Req.UserInfos.Userip}}</req:userip>
            <req:userport>{{.Req.UserInfos.Userport}}</req:userport>
            <req:username>{{.Req.UserInfos.Username}}</req:username>
          </sp:item>
        </req:userInfos>
      </sp:req>
    </sp:{{.Method}}>
  </x:Body>
</x:Envelope>`

	buf := bytes.NewBufferString("")
	t, _ := template.New("a").Parse(soapBody)
	isStart := false
	if method == "startQos" {
		isStart = true
	}
	t.Execute(buf, NewData{Req: data, Method: method, IsStart: isStart})

	return buf.String()
}

func createUserSOAPXml(productid string, ip string) string {
	soapBody := `
<x:Envelope xmlns:x="http://schemas.xmlsoap.org/soap/envelope/">
  <x:Header/>
  <x:Body>
    <getUserProduct xmlns="">
      <productid>{{.Productid}}</productid>
      <ip>{{.Ip}}</ip>
    </getUserProduct>
  </x:Body>
</x:Envelope>`

	buf := bytes.NewBufferString("")
	t, _ := template.New("a").Parse(soapBody)
	t.Execute(buf, struct {
		Productid string
		Ip        string
	}{productid, ip})

	return buf.String()
}

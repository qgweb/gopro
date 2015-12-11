// 具体业务罗辑
package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/qgweb/gopro/qianzhao-ht-tcpserver/common/function"
	"regexp"
	"strings"
	"text/template"
	"time"

	"github.com/qgweb/gopro/lib/convert"

	"github.com/astaxie/beego/httplib"
	"github.com/ngaut/log"
	"github.com/qgweb/gopro/lib/encrypt"
	"github.com/qgweb/gopro/qianzhao-ht-tcpserver/model"
)

const (
	APP_KEY        = "APP_MSB704PU"
	APP_SECRET     = "mv6oy8f2qo0l0ogvxnm02tM7"
	EBIT_BASE_URL  = "http://218.85.118.9:8000/api2/"
	CARD_APPLY_URL = "http://221.228.17.114:58886/iherbhelper/services/ApplicationCard"
)

// 海淘卡信息
type Card struct {
	Flag          string `json:"Flag"`
	CardID        string `json:"CardID"`
	CardPass      string `json:"CardPass"`
	Serviceid     string `json:"Serviceid"`
	Transactionid string `json:"Transactionid"`
}

type BDInterfaceManager struct{}

func (this *BDInterfaceManager) HaveTime(account string) int {
	model := model.BrandAccountRecord{}
	return model.GetAccountCanUserTime(account)
}

// 判断是否开启
func (this *BDInterfaceManager) CanStart(ip string) string {
	var (
		timestamp = convert.ToString(time.Now().Unix())
		secret    = getSecret(timestamp)
	)

	// 正式删除
	//return "1111"

	req := httplib.Post(EBIT_BASE_URL + "user/query")
	req.JsonBody(map[string]string{
		"app":       APP_KEY,
		"secret":    secret,
		"timestamp": timestamp,
		"_type":     "0",
		"data":      ip,
	})

	res := make(map[string]interface{})
	req.ToJson(&res)
	if CheckError(res) {
		return ""
	}

	if task_id, ok := res["task_id"]; ok {
		time.Sleep(time.Second * 3)
		info := TaskQuery(task_id.(string))
		if CheckError(info) {
			return ""
		}
		return info["dial_acct"].(string)
	}

	return ""

}

// 开启
func (this *BDInterfaceManager) Start(account string, ip string) Respond {
	var (
		timestamp = convert.ToString(time.Now().Unix())
		secret    = getSecret(timestamp)
		errmsg    = "抱歉，程序发生错误,提速失败"
	)

	//// 正式删除
	//return Respond{Code: "200", Msg: "xxxxxxxxxxx"}

	req := httplib.Post(EBIT_BASE_URL + "speedup/open")
	req.JsonBody(map[string]string{
		"app":       APP_KEY,
		"secret":    secret,
		"timestamp": timestamp,
		"ip_addr":   ip,
		"duration":  "60",
		"dial_acct": account,
	})
	res := make(map[string]interface{})
	req.ToJson(&res)

	if CheckError(res) {
		return Respond{Code: "500", Msg: errmsg}
	}

	if task_id, ok := res["task_id"]; ok {
		time.Sleep(time.Second * 3)
		info := TaskQuery(task_id.(string))
		if CheckError(info) {
			return Respond{Code: "500", Msg: errmsg}
		}
		return Respond{Code: "200", Msg: info["channel_id"].(string)}
	}

	return Respond{Code: "500", Msg: errmsg}
}

// 关闭
func (this *BDInterfaceManager) Stop(channel_id string) Respond {
	var (
		timestamp = convert.ToString(time.Now().Unix())
		secret    = getSecret(timestamp)
		errmsg    = "抱歉，程序发生错误,提速关闭失败"
	)
	//// 正式删除
	//return Respond{Code: "200", Msg: ""}

	req := httplib.Post(EBIT_BASE_URL + "speedup/close")
	req.JsonBody(map[string]string{
		"app":        APP_KEY,
		"secret":     secret,
		"timestamp":  timestamp,
		"channel_id": channel_id,
	})
	res := make(map[string]interface{})
	req.ToJson(&res)

	if CheckError(res) {
		return Respond{Code: "500", Msg: errmsg}
	}

	if task_id, ok := res["task_id"]; ok {
		time.Sleep(time.Second * 3)
		info := TaskQuery(task_id.(string))
		if CheckError(info) {
			return Respond{Code: "500", Msg: errmsg}
		}
		return Respond{Code: "200", Msg: ""}
	}
	return Respond{Code: "500", Msg: errmsg}
}

func CheckError(res map[string]interface{}) bool {
	if err, ok := res["errno"]; ok && err.(float64) != 0 {
		if msg, ok := res["message"]; ok {
			log.Error(msg)
		}
		return true
	}
	return false
}

func TaskQuery(taskId string) map[string]interface{} {
	var (
		timestamp = convert.ToString(time.Now().Unix())
		secret    = getSecret(timestamp)
	)

	req := httplib.Post(EBIT_BASE_URL + "task/query")

	body := make(map[string]string)
	body["app"] = APP_KEY
	body["secret"] = secret
	body["timestamp"] = timestamp
	body["task_id"] = taskId

	req.JsonBody(&body)

	v := make(map[string]interface{})

	req.ToJson(&v)
	return v
}

func getSecret(timestamp string) string {
	return encrypt.DefaultMd5.Encode(APP_KEY + timestamp + APP_SECRET)
}

//1、正式接口地址为
//http://202.102.13.123:17001/services/LcimsForUserInfo?wsdl
//
//他们可以使用202.102.41.115来调用。
//
//2、测试接口地址为
//http://202.102.13.98:7001/services/LcimsForUserInfo?wsdl
//
//他们可以用  122.225.98.80 （浙江测试server） 和202.102.41.115（江苏云平台  正式server） 来调用。

//http://221.228.17.114:58886/iherbhelper/services/ApplicationCard?wsdl

// serviceid 0001

func (this *BDInterfaceManager) FreeCardApplyFor(phone string) (card Card) {
	//测试
	//{"Flag":"0","CardID":"56000005038843","CardPass":"zta3t7M0","Serviceid":"0001","Transactionid":"1449817500610"}
	//{0 56000005039489 WhVgF3cR 0001 1449819617343}
	card.Flag = "0"
	card.CardID = "56000005038843"
	card.CardPass = "zta3t7M0"
	card.Serviceid = "0001"
	card.Transactionid = "1449817500610"
	return card

	//base64 (AES(JSON(Mobile ,Serviceid)))
	//MD5（Mobile（未加密）+ ENCRYPT）.toUpperCase
	var tmp = `<x:Envelope xmlns:x="http://schemas.xmlsoap.org/soap/envelope/" xmlns:ser="http://service.com">
    <x:Header/>
    <x:Body>
        <ser:applicationCard>
            <ser:in0>{{.Phone}}</ser:in0>
            <ser:in1>{{.Encrypt}}</ser:in1>
            <ser:in2>{{.Digest}}</ser:in2>
        </ser:applicationCard>
    </x:Body>
</x:Envelope>`
	var serviceid = "0001"
	var jsonMobil = fmt.Sprintf(`{"Mobile":"%s","Serviceid":"%s"}`, phone, serviceid)
	log.Info(jsonMobil)
	log.Info(function.AESEncrypt(jsonMobil))
	var encryptVal = function.AESEncrypt(jsonMobil)
	var digest = strings.ToUpper(encrypt.DefaultMd5.Encode(phone + encryptVal))

	t, _ := template.New("").Parse(tmp)
	bf := bytes.NewBuffer(nil)
	t.Execute(bf, struct {
		Phone   string
		Encrypt string
		Digest  string
	}{phone, encryptVal, digest})

	req := httplib.Post(CARD_APPLY_URL)
	req.Header("Content-Type", "text/xml; charset=utf-8")
	req.Header("SOAPAction", CARD_APPLY_URL)
	req.Body(bf.String())
	log.Error(bf.String())
	resp, err := req.String()
	if err != nil {
		log.Error(err)
		return card
	}

	//resp := `<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/" xmlns:xsd="http://www.w3.org/2001/XMLSchema" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"><soap:Body><ns1:applicationCardResponse xmlns:ns1="http://service.com"><ns1:out>MA5I8gcCKyVzZ+CJdJnsklJDdf2X6oVyoSxBvuVKA7cdQ+BkxCGWQPEiB0qNo+iAgkwceu6kIx2gBHTVSW6gipTQRgj9Ck1fWGhTH55z5Vj5KL6g9quokror6EsjZYLT/SrQ4fkC7r4E8F8qUJOgDw==</ns1:out></ns1:applicationCardResponse></soap:Body></soap:Envelope>`
	r, _ := regexp.Compile(`<ns1:out>(.+)</ns1:out>`)
	if vv := r.FindStringSubmatch(resp); len(vv) >= 2 {
		log.Info(function.AESDecrypt(strings.TrimSpace(vv[1])))
		err := json.Unmarshal([]byte(function.AESDecrypt(strings.TrimSpace(vv[1]))), &card)
		if err != nil {
			log.Error(vv[1])
			log.Error(err)
		}
	}
	return
}

/*
<!--?xml version="1.0" encoding="utf-8" ?-->
<soapenv:envelope xmlns:soapenv="http://schemas.xmlsoap.org/soap/envelope/" xmlns:xsd="http://www.w3.org/2001/XMLSchema" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
    <soapenv:body>
        <getcarduserinforesponse soapenv:encodingstyle="http://schemas.xmlsoap.org/soap/encoding/">
            <getcarduserinforeturn href="#id0"></getcarduserinforeturn>
        </getcarduserinforesponse>
        <multiref id="id0" soapenc:root="0" soapenv:encodingstyle="http://schemas.xmlsoap.org/soap/encoding/" xsi:type="ns1:GetCardUserInfoResponse" xmlns:soapenc="http://schemas.xmlsoap.org/soap/encoding/" xmlns:ns1="http://lcimsforuserinfo.webservice.lcbms.linkage.com">
            <cardinfoxml xsi:type="soapenc:string" xsi:nil="true"></cardinfoxml>
            <errordescription xsi:type="soapenc:string">用户不存在!</errordescription>
            <result href="#id1"></result>
        </multiref>
        <multiref id="id1" soapenc:root="0" soapenv:encodingstyle="http://schemas.xmlsoap.org/soap/encoding/" xsi:type="xsd:int" xmlns:soapenc="http://schemas.xmlsoap.org/soap/encoding/">-1</multiref>
    </soapenv:body>
</soapenv:envelope>
*/

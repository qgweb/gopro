// 具体业务罗辑
package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/qgweb/gopro/qianzhao-ht-tcpserver/common/function"
	"regexp"
	"strings"
	"text/template"

	"github.com/qgweb/gopro/lib/convert"

	"errors"
	"github.com/astaxie/beego/httplib"
	"github.com/ngaut/log"
	"github.com/qgweb/gopro/lib/encrypt"
	"github.com/qgweb/gopro/qianzhao-ht-tcpserver/model"
	"github.com/qgweb/gopro/qianzhao/common/config"
)

const (
	APP_KEY             = "APP_MSB704PU"
	APP_SECRET          = "mv6oy8f2qo0l0ogvxnm02tM7"
	EBIT_BASE_URL       = "http://218.85.118.9:8000/api2/"
	CARD_APPLY_URL      = "http://221.228.17.114:58886/iherbhelper/services/ApplicationCard"
	CARD_CHECK_URL      = "http://221.228.17.114:58886/iherbhelper/services/CheckCard"
	CARD_QUERY_TEST_URL = "http://202.102.13.98:7001/services/LcimsForUserInfo"
	CARD_QUERY_URL      = "http://202.102.13.123:7001/services/LcimsForUserInfo"
	FREE_CARD           = 0
	UNFREE_CARD         = 1
)

type Card struct {
	Flag          string `json:"Flag"`
	CardID        string `json:"CardID"`
	CardPass      string `json:"CardPass"`
	Serviceid     string `json:"Serviceid"`
	Transactionid string `json:"Transactionid"`
}

type MCard struct {
	CardNO    string `json:"CardNO"`
	CardPass  string `json:"CardPass"`
	Serviceid string `json:"Serviceid"`
	Digest    string `json:"Digest"`
	Mobile    string `json:"Mobile"`
}

type BDInterfaceManager struct{}

func (this *BDInterfaceManager) GetQueryUrl() string {
	if v,_:=config.GetDefault().Key("debug").Int();v==1 {
		return CARD_QUERY_TEST_URL
	}
	return CARD_QUERY_URL
}

// 开启
func (this *BDInterfaceManager) Start(card MCard, cardType int) Respond {
	log.Info(config.GetDefault().Key("debug").Int())
	if cardType == 0 { //免费卡
		ht, err := this.freeCard(card.Mobile)
		r := Respond{}
		if err != nil {
			r.Code = "500"
			r.Msg = err.Error()
			return r
		} else {
			r.Code = "200"
			r.Msg = fmt.Sprintf("%d|%s|%s|%s|%d", ht.Id, ht.CardNum, ht.CardPwd, ht.CardToken, ht.TotalTime)
			return r
		}

	}
	if cardType == 1 { //购买卡
		ht, err := this.moneyCard(card)
		r := Respond{}
		if err != nil {
			r.Code = "500"
			r.Msg = err.Error()
			return r
		} else {
			r.Code = "200"
			r.Msg = fmt.Sprintf("%d|%s|%s|%s|%d", ht.Id, ht.CardNum, ht.CardPwd, ht.CardToken, ht.TotalTime)
			return r
		}
	}
	return Respond{"500", "系统发生错误"}
}

func (this *BDInterfaceManager) freeCard(phone string) (hcard model.HTCard, err error) {
	var (
		hmodel  = model.HTCard{}
		date    = function.GetDateUnix()
	)
	ht := hmodel.GetInfoByPhone(phone, date, 0, 1)
	if ht.Id == 0 {
		//申请卡
		card := this.FreeCardApplyFor(phone)
		if card.Flag == "-2" {
			return hcard, errors.New("体验卡已经发放完毕")
		}
		if card.Flag != "0" {
			return hcard, errors.New("系统发生异常")
		}
		info := model.HTCard{}
		info.CardNum = card.CardID
		info.CardPwd = card.CardPass
		info.CardToken = card.Transactionid
		info.CardType = 0
		info.Date = convert.ToInt(date)
		info.Phone = phone
		info.Remark = ""
		info.Status = 1
		info.TotalTime = model.HT_SPEED_UP_TIME
		info.Id = hmodel.AddReocrd(info)
		return info, nil
	} else {
		balance := this.CardInfoQuery(ht.CardNum)
		if balance <= 0 {
			ht.Status = 3
			hmodel.UpdateCard(ht)
			ht.TotalTime = 0
			return ht, errors.New("用户免费体验时间已到")
		}
		return ht, nil
	}
}

func (this *BDInterfaceManager) moneyCard(card MCard) (hcard model.HTCard, err error) {
	var (
		hmodel  = model.HTCard{}
		date    = function.GetDateUnix()
	)

	//验证卡有效
	tid, err := this.CheckCardCanUse(card)
	log.Error(err)
	if err != nil {
		log.Error(err)
		return hcard, err
	}

	ht := hmodel.GetInfoByCard(card.Mobile, date, card.CardNO, 1)
	balance := this.CardInfoQuery(card.CardNO)

	if ht.Id == 0 && balance > 0 {
		info := model.HTCard{}
		info.CardNum = card.CardNO
		info.CardPwd = card.CardPass
		info.CardToken = tid
		info.CardType = 1
		info.Date = convert.ToInt(date)
		info.Phone = card.Mobile
		info.Remark = ""
		info.Status = 1
		info.TotalTime = balance * 72
		ht.Id = hmodel.AddReocrd(info)
		return info, nil
	} else {
		if balance <= 0 {
			ht.Status = 3
			hmodel.UpdateCard(ht)
			ht.TotalTime = 0
			return ht, errors.New("用户免费体验时间已到")
		}
	}
	return hcard,nil
}

// 关闭
func (this *BDInterfaceManager) Stop(channel_id string) Respond {
	return Respond{}
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
		log.Info(card)
		if err != nil {
			log.Error(vv[1])
			log.Error(err)
		}
	}
	return
}

func (this *BDInterfaceManager) CheckCardCanUse(card MCard) (string, error) {
	var tmp = `<x:Envelope xmlns:x="http://schemas.xmlsoap.org/soap/envelope/" xmlns:ser="http://service.com">
    <x:Header/>
    <x:Body>
        <ser:checkCard>
            <ser:in0>{{.CardNO}}</ser:in0>
			<ser:in1>{{.CardPass}}</ser:in1>
			<ser:in2>{{.Serviceid}}</ser:in2>
			<ser:in3>{{.Mobile}}</ser:in3>
			<ser:in4>{{.Digest}}</ser:in4>
        </ser:checkCard>
    </x:Body>
</x:Envelope>`

	card.CardNO = function.AESEncrypt(card.CardNO)
	card.CardPass = function.AESEncrypt(card.CardPass)
	card.Serviceid = function.AESEncrypt(card.Serviceid)
	card.Digest = strings.ToUpper(encrypt.DefaultMd5.Encode(card.CardNO + card.CardPass + card.Serviceid + card.Mobile))

	bf := bytes.NewBuffer(nil)
	t, _ := template.New("").Parse(tmp)
	t.Execute(bf, card)

	req := httplib.Post(CARD_CHECK_URL)
	req.Header("Content-Type", "text/xml; charset=utf-8")
	req.Header("SOAPAction", CARD_CHECK_URL)
	req.Body(bf.String())
	log.Error(bf.String())
	resp, err := req.String()
	if err != nil {
		log.Error(err)
		return "", errors.New("系统异常")
	}

	//resp := `<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/" xmlns:xsd="http://www.w3.org/2001/XMLSchema" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"><soap:Body>  <ns1:checkCardResponse xmlns:ns1="http://service.com"><ns1:out>aJaWUxQpbge+M2JLSMQLWg==</ns1:out></ns1:checkCardResponse></soap:Body></soap:Envelope>`
	r, _ := regexp.Compile(`<ns1:out>(.+)</ns1:out>`)
	if vv := r.FindStringSubmatch(resp); len(vv) >= 2 {
		log.Info(function.AESDecrypt(strings.TrimSpace(vv[1])))
		flag := make(map[string]interface{})
		err := json.Unmarshal([]byte(function.AESDecrypt(strings.TrimSpace(vv[1]))), &flag)
		log.Info(flag)
		if err != nil {
			log.Error(vv[1])
			log.Error(err)
			return "", errors.New("系统异常")
		}

		/*
			0：成功
			-1 :不存在账号
			-2：密码错误
			-3：该卡类型不对
			-4：该卡正在使用中
			-10：参数格式错误
			-11：参数校验失败
			-99：系统异常
		*/
		switch flag["Flag"].(string) {
		case "0","-4":
			return flag["Transactionid"].(string), nil
		case "-1":
			return "", errors.New("不存在账号")
		case "-2":
			return "", errors.New("密码错误")
		case "-3", "-10", "-99", "-11":
			return "", errors.New("系统异常")
		}
	}
	return "", errors.New("系统异常")
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

// 查询卡余额信息
func (this *BDInterfaceManager) CardInfoQuery(cardNo string) int {
	var rhead = `<soapenv:Envelope xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xmlns:xsd="http://www.w3.org/2001/XMLSchema" xmlns:soapenv="http://schemas.xmlsoap.org/soap/envelope/" xmlns:lcim="http://lcimsforuserinfo.webservice.lcbms.linkage.com">
   <soapenv:Header/>
   <soapenv:Body>
      <lcim:getCardUserInfo soapenv:encodingStyle="http://schemas.xmlsoap.org/soap/encoding/">
         <in0 xsi:type="lcim:GetCardUserInfoRequest">
            <cardno xsi:type="soapenc:string" xmlns:soapenc="http://schemas.xmlsoap.org/soap/encoding/">` + cardNo + `</cardno>
         </in0>
      </lcim:getCardUserInfo>
   </soapenv:Body>
</soapenv:Envelope>`
	req := httplib.Post(this.GetQueryUrl())
	req.Header("Content-Type", "text/xml; charset=utf-8")
	req.Header("SOAPAction", this.GetQueryUrl())
	req.Body(rhead)
	body, err := req.String()
	if err != nil {
		log.Error(err)
		return 0
	}

	p, err := goquery.NewDocumentFromReader(strings.NewReader(body))
	if err != nil {
		log.Error(err)
		log.Error(body)
		return 0
	}
	balance := p.Find("cardinfoxml").Find("balance").Text()
	if balance == "" {
		return 0
	}
	// 200分->2小时
	// 1分->72秒
	return convert.ToInt(balance)
}

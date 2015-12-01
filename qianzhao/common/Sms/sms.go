// 短信接口
package Sms

import (
	"bytes"
	"github.com/astaxie/beego/httplib"
	"log"
	"text/template"
)

/*
账号：025C86588423
密码：86588423
计费号码：02586588423

其他参数参照接口文档

先给开了15000条/月
*/

const (
	SMS_ACCOUNT  = "025C86588423"
	SMS_PWD      = "86588423"
	SMS_SEND_URL = "http://202.102.41.99:8090/wsewebsm/services/SendMessageService"
)

func createBody(phone string, content string) string {
	var body = `<soapenv:Envelope 
xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xmlns:xsd="http://www.w3.org/2001/XMLSchema" 	xmlns:soapenv="http://schemas.xmlsoap.org/soap/envelope/" 	xmlns:sen="http://sendmessage.service.webservice.linkage.com">
   <soapenv:Header/>
   <soapenv:Body>
      <sen:sendSms soapenv:encodingStyle="http://schemas.xmlsoap.org/soap/encoding/">
         <simpleUserInfo  xsi:type="sen:SimpleUserInfo" 
xmlns:sen="http://sendmessage.server.webservice.linkage.com">
             <username  xsi:type="soapenc:string" 
xmlns:soapenc="http://schemas.xmlsoap.org/soap/encoding/">{{.Account}}</username>
  <password  xsi:type="soapenc:string" 
xmlns:soapenc="http://schemas.xmlsoap.org/soap/encoding/">{{.Pwd}}</password>
         </simpleUserInfo>
         <sendSmsInfo xsi:type="sen:SendSmsRequest" 
xmlns:sen="http://sendmessage.server.webservice.linkage.com">
            <content  xsi:type="soapenc:string" 
xmlns:soapenc="http://schemas.xmlsoap.org/soap/encoding/">{{.Content}}</content>
            <receiveNum  xsi:type="soapenc:string" 
xmlns:soapenc="http://schemas.xmlsoap.org/soap/encoding/">{{.Phone}}</receiveNum>
            <sendType  xsi:type="soapenc:string" 
xmlns:soapenc="http://schemas.xmlsoap.org/soap/encoding/">1</sendType>
            <signature xsi:type="soapenc:string" 
xmlns:soapenc="http://schemas.xmlsoap.org/soap/encoding/"></signature>
         </sendSmsInfo>
      </sen:sendSms>
   </soapenv:Body>
</soapenv:Envelope>`

	type SmsContent struct {
		Account string
		Pwd     string
		Content string
		Phone   string
	}

	bf := bytes.NewBuffer(nil)
	sc := SmsContent{}
	sc.Account = SMS_ACCOUNT
	sc.Pwd = SMS_PWD
	sc.Content = content
	sc.Phone = phone
	t, err := template.New("xx").Parse(body)
	if err == nil {
		log.Println(t.Execute(bf, sc))
	}
	return bf.String()
}

// 发送短信
func SendMsg(phone string, content string) {
	body := createBody(phone, content)
	log.Println(body)
	h := httplib.Post(SMS_SEND_URL)
	h.Header("SOAPAction", SMS_SEND_URL)
	h.Body(body)
	log.Println(h.String())
}

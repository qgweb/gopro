package main

import (
	"flag"
	"fmt"
	"github.com/astaxie/beego/httplib"
	// "github.com/donnie4w/dom4g"
	//"github.com/opesun/goquery"
)

const (
	QOS_QUERY_URL = "http://202.102.13.98:7001/services/LcimsForUserInfo"
)

func main() {
	var num = flag.String("n", "", "卡号")
	flag.Parse()
	var query = `<soapenv:Envelope xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xmlns:xsd="http://www.w3.org/2001/XMLSchema" xmlns:soapenv="http://schemas.xmlsoap.org/soap/envelope/" xmlns:lcim="http://lcimsforuserinfo.webservice.lcbms.linkage.com">
   <soapenv:Header/>
   <soapenv:Body>
      <lcim:getCardUserInfo soapenv:encodingStyle="http://schemas.xmlsoap.org/soap/encoding/">
         <in0 xsi:type="lcim:GetCardUserInfoRequest">
            <cardno xsi:type="soapenc:string" xmlns:soapenc="http://schemas.xmlsoap.org/soap/encoding/">` + *num + `</cardno>
         </in0>
      </lcim:getCardUserInfo>
   </soapenv:Body>
</soapenv:Envelope>`
	fmt.Println(query)
	req := httplib.Post(QOS_QUERY_URL)
	req.Header("Content-Type", "text/xml; charset=utf-8")
	req.Header("SOAPAction", QOS_QUERY_URL)
	req.Body(query)
	fmt.Println(req.String())
}

/**
 * <!--?xml version="1.0" encoding="utf-8" ?-->
<soapenv:envelope xmlns:soapenv="http://schemas.xmlsoap.org/soap/envelope/" xmlns:xsd="http://www.w3.org/2001/XMLSchema" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
    <soapenv:body>
        <ns1:getcarduserinforesponse soapenv:encodingstyle="http://schemas.xmlsoap.org/soap/encoding/" xmlns:ns1="http://lcimsforuserinfo.webservice.lcbms.linkage.com">
            <getcarduserinforeturn href="#id0"></getcarduserinforeturn>
        </ns1:getcarduserinforesponse>
        <multiref id="id0" soapenc:root="0" soapenv:encodingstyle="http://schemas.xmlsoap.org/soap/encoding/" xsi:type="ns2:GetCardUserInfoResponse" xmlns:soapenc="http://schemas.xmlsoap.org/soap/encoding/" xmlns:ns2="http://lcimsforuserinfo.webservice.lcbms.linkage.com">
            <cardinfoxml xsi:type="soapenc:string">
            <?xml version="1.0" encoding="gb2312"?>
            	<cardinfo><cardno>56000005038843</cardno>
            	<cardtype>海淘上网助手</cardtype>
            	<cardbat>S00003703</cardbat>
            	<groupiddes>DSL拨号-海淘助手</groupiddes>
            	<cardvalue>200</cardvalue>
            	<balance>200</balance>
            	<areano>省中心</areano>
            	<limitusers>1</limitusers>
            	<opendate>2015-12-14 11:48:33</opendate>
            	<effectdate>2015-12-14 11:48:33</effectdate>
            	<expiredate>2015-12-30 23:59:05</expiredate>
            	<status>0</status></cardinfo>
            	</cardinfoxml>
            <errordescription xsi:type="soapenc:string">卡用户基本信息查询成功!</errordescription>
            <result href="#id1"></result>
        </multiref>
        <multiref id="id1" soapenc:root="0" soapenv:encodingstyle="http://schemas.xmlsoap.org/soap/encoding/" xsi:type="xsd:int" xmlns:soapenc="http://schemas.xmlsoap.org/soap/encoding/">0</multiref>
    </soapenv:body>
</soapenv:envelope>
*/

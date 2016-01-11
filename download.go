package main

import (
	"github.com/PuerkitoBio/goquery"
	"strings"
	"github.com/ngaut/log"
)

func main() {
	str := `<?xml version="1.0" encoding="utf-8"?><soapenv:Envelope xmlns:soapenv="http://schemas.xmlsoap.org/soap/envelope/" xmlns:xsd="http://www.w3.org/2001/XMLSchema" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"><soapenv:Body><ns1:getCardUserInfoResponse soapenv:encodingStyle="http://schemas.xmlsoap.org/soap/encoding/" xmlns:ns1="http://lcimsforuserinfo.webservice.lcbms.linkage.com"><getCardUserInfoReturn href="#id0"/></ns1:getCardUserInfoResponse><multiRef id="id0" soapenc:root="0" soapenv:encodingStyle="http://schemas.xmlsoap.org/soap/encoding/" xsi:type="ns2:GetCardUserInfoResponse" xmlns:soapenc="http://schemas.xmlsoap.org/soap/encoding/" xmlns:ns2="http://lcimsforuserinfo.webservice.lcbms.linkage.com"><cardInfoXML xsi:type="soapenc:string">&lt;?xml version=&quot;1.0&quot; encoding=&quot;gb2312&quot;?&gt;&lt;cardinfo&gt;&lt;cardno&gt;56000005038843&lt;/cardno&gt;&lt;cardtype&gt;&#x6D77;&#x6DD8;&#x4E0A;&#x7F51;&#x52A9;&#x624B;&lt;/cardtype&gt;&lt;cardbat&gt;S00003709&lt;/cardbat&gt;&lt;groupiddes&gt;DSL&#x62E8;&#x53F7;-&#x6D77;&#x6DD8;&#x52A9;&#x624B;&lt;/groupiddes&gt;&lt;cardvalue&gt;200&lt;/cardvalue&gt;&lt;balance&gt;200&lt;/balance&gt;&lt;areano&gt;&#x7701;&#x4E2D;&#x5FC3;&lt;/areano&gt;&lt;limitusers&gt;1&lt;/limitusers&gt;&lt;opendate&gt;2015-10-26 16:14:12&lt;/opendate&gt;&lt;effectdate&gt;2015-12-22 16:49:52&lt;/effectdate&gt;&lt;expiredate&gt;2016-01-21 16:49:52&lt;/expiredate&gt;&lt;status&gt;0&lt;/status&gt;&lt;/cardinfo&gt;</cardInfoXML><errorDescription xsi:type="soapenc:string">&#x5361;&#x7528;&#x6237;&#x57FA;&#x672C;&#x4FE1;&#x606F;&#x67E5;&#x8BE2;&#x6210;&#x529F;!</errorDescription><result href="#id1"/></multiRef><multiRef id="id1" soapenc:root="0" soapenv:encodingStyle="http://schemas.xmlsoap.org/soap/encoding/" xsi:type="xsd:int" xmlns:soapenc="http://schemas.xmlsoap.org/soap/encoding/">0</multiRef></soapenv:Body></soapenv:Envelope>`
	p, err := goquery.NewDocumentFromReader(strings.NewReader(str))
	if err != nil {
		log.Error(err)
		return
	}


	balance,err := goquery.NewDocumentFromReader(strings.NewReader(p.Find("cardinfoxml").Text()));
	if err !=nil {

	}
	log.Info(balance.Find("balance").Text())

}

// package main

// //可能有多余的导入包
// import (
// 	"bytes"
// 	"fmt"
// 	"github.com/astaxie/beego/httplib"
// 	"io/ioutil"
// 	"net/http"
// )

// //POST到webService
// func PostWebService(url string, method string, value string) string {
// 	res, err := http.Post(url, "application/soap+xml; charset=utf-8", bytes.NewBuffer([]byte(value)))
// 	//这里随便传递了点东西
// 	if err != nil {
// 		fmt.Println("post error", err)
// 	}
// 	data, err := ioutil.ReadAll(res.Body)
// 	//取出主体的内容
// 	if err != nil {
// 		fmt.Println("read error", err)
// 	}
// 	res.Body.Close()
// 	fmt.Printf("result----%s", data)
// 	return string(data)
// }

// func CreateSOAPXml(nameSpace string, methodName string, productid string, ip string) string {
// 	soapBody := "<?xml version=\"1.0\" encoding=\"utf-8\"?>"
// 	soapBody += "<soap12:Envelope xmlns:m=\"http://js.vnet.cn\" xmlns:xsi=\"http://www.w3.org/2001/XMLSchema-instance\" xmlns:xsd=\"http://www.w3.org/2001/XMLSchema\" xmlns:soap12=\"http://www.w3.org/2003/05/soap-envelope\">"
// 	soapBody += "<soap12:Body>"
// 	soapBody += "<" + methodName + " xmlns=\"" + nameSpace + "\">"
// 	//以下是具体参数
// 	soapBody += "<productid>" + productid + "</productid>"
// 	soapBody += "<ip>" + ip + "</ip>"
// 	soapBody += "</" + methodName + "></soap12:Body></soap12:Envelope>"
// 	return soapBody
// }

// func main() {
// 	postStr := CreateSOAPXml("http://tempuri.org/", "getUserProduct", "192.168.1.199:33", "1100000300002204")
// 	r := httplib.Post("http://js.vnet.cn/ProvinceForSPSearchUserName/services/ProvinceForSPSearchUserName?wsdl")
// 	r.Header("Content-Type", "text/xml; charset=utf-8")
// 	r.Header("SOAPAction", "http://js.vnet.cn/ProvinceForSPSearchUserName/services/ProvinceForSPSearchUserName")
// 	r.Body(postStr)
// 	fmt.Println(r.String())
// }

package main

import (
	"encoding/xml"
	"fmt"
)

// <soapenv:Envelope xmlns:soapenv="http://www.w3.org/2003/05/soap-envelope" xmlns:xsd="http://www.w3.org/2001/XMLSchema" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
//   <soapenv:Body>
//     <ns1:getUserProductResponse xmlns:ns1="http://tempuri.org/" soapenv:encodingStyle="http://www.w3.org/2003/05/soap-encoding">
//       <ns2:result xmlns:ns2="http://www.w3.org/2003/05/soap-rpc">getUserProductReturn</ns2:result>
//       <getUserProductReturn xmlns:ns3="http://schemas.xmlsoap.org/soap/encoding/" xsi:type="ns3:string">-998||</getUserProductReturn>
//     </ns1:getUserProductResponse>
//   </soapenv:Body>
// </soapenv:Envelope>

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

func main() {

	input := `<?xml version="1.0" encoding="utf-8"?><soapenv:Envelope xmlns:soapenv="http://www.w3.org/2003/05/soap-envelope" xmlns:xsd="http://www.w3.org/2001/XMLSchema" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"><soapenv:Body><ns1:getUserProductResponse soapenv:encodingStyle="http://www.w3.org/2003/05/soap-encoding" xmlns:ns1="http://tempuri.org/"><ns2:result xmlns:ns2="http://www.w3.org/2003/05/soap-rpc">getUserProductReturn</ns2:result><getUserProductReturn xsi:type="ns3:string" xmlns:ns3="http://schemas.xmlsoap.org/soap/encoding/">-998||</getUserProductReturn></ns1:getUserProductResponse></soapenv:Body></soapenv:Envelope>`
	//inputReader := strings.NewReader(input)

	// 从文件读取，如可以如下：
	// content, err := ioutil.ReadFile("studygolang.xml")
	// decoder := xml.NewDecoder(bytes.NewBuffer(content))
	var by Envelope
	fmt.Println(xml.Unmarshal([]byte(input), &by))
	fmt.Println(by.By.GP.Result)
}

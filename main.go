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

// package main

// import (
// 	"encoding/xml"
// 	"fmt"
// )

// // <soapenv:Envelope xmlns:soapenv="http://www.w3.org/2003/05/soap-envelope" xmlns:xsd="http://www.w3.org/2001/XMLSchema" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
// //   <soapenv:Body>
// //     <ns1:getUserProductResponse xmlns:ns1="http://tempuri.org/" soapenv:encodingStyle="http://www.w3.org/2003/05/soap-encoding">
// //       <ns2:result xmlns:ns2="http://www.w3.org/2003/05/soap-rpc">getUserProductReturn</ns2:result>
// //       <getUserProductReturn xmlns:ns3="http://schemas.xmlsoap.org/soap/encoding/" xsi:type="ns3:string">-998||</getUserProductReturn>
// //     </ns1:getUserProductResponse>
// //   </soapenv:Body>
// // </soapenv:Envelope>

// type getUserProductResponse struct {
// 	Result               string `xml:"result"`
// 	GetUserProductReturn string `xml:"getUserProductReturn"`
// }

// type Envelope struct {
// 	By Body `xml:"Body"`
// }

// type Body struct {
// 	GP getUserProductResponse `xml:"getUserProductResponse"`
// }

// func main() {

// 	input := `<?xml version="1.0" encoding="utf-8"?><soapenv:Envelope xmlns:soapenv="http://www.w3.org/2003/05/soap-envelope" xmlns:xsd="http://www.w3.org/2001/XMLSchema" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"><soapenv:Body><ns1:getUserProductResponse soapenv:encodingStyle="http://www.w3.org/2003/05/soap-encoding" xmlns:ns1="http://tempuri.org/"><ns2:result xmlns:ns2="http://www.w3.org/2003/05/soap-rpc">getUserProductReturn</ns2:result><getUserProductReturn xsi:type="ns3:string" xmlns:ns3="http://schemas.xmlsoap.org/soap/encoding/">-998||</getUserProductReturn></ns1:getUserProductResponse></soapenv:Body></soapenv:Envelope>`
// 	//inputReader := strings.NewReader(input)

// 	// 从文件读取，如可以如下：
// 	// content, err := ioutil.ReadFile("studygolang.xml")
// 	// decoder := xml.NewDecoder(bytes.NewBuffer(content))
// 	var by Envelope
// 	fmt.Println(xml.Unmarshal([]byte(input), &by))
// 	fmt.Println(by.By.GP.Result)
// }

//package main

//import (
//	//"errors"
//	//"github.com/astaxie/beego/logs"
//	"github.com/juju/errors"
//	"github.com/ngaut/log"
//)

//func show1() error {
//	return errors.Trace(errors.New("aaa"))
//}

//func show() error {
//	return errors.Annotate(errors.Trace(show1()), "bbbbb")
//}

//func main() {
//	//log.SetOutputByName("./b.txt")
//	log.SetRotateByHour()
//	log.SetHighlighting(true)
//	log.SetLevelByString("errors")
//	log.SetFlags(log.Lshortfile | log.Ltime | log.Ldate)
//	log.Warn("ffffff")

//	// log := logs.NewLogger(10000)
//	// log.SetLogger("console", "")
//	// log.EnableFuncCallDepth(true)
//	// log.Debug("debug")
//	//
//	log.Info(errors.ErrorStack(show()))
//}
//package main

//import (
//	"bytes"
//	"fmt"
//	"net"
//	"time"

//	. "code.google.com/p/go.crypto/ssh"
//	"gopkg.in/mgo.v2"
//	"gopkg.in/mgo.v2/bson"
//)

//func main() {
//	config := &ClientConfig{
//		User: "root",
//		Auth: []AuthMethod{
//			Password("a77da56syDdDLNiW"),
//		},
//	}
//	client, err := Dial("tcp", "122.225.98.68:22", config)
//	if err != nil {
//		panic("Failed to dial: " + err.Error())
//	}

//	sess, err := getMongo("xu:xu123net@192.168.0.68:10001/xu_precise", func() (net.Conn, error) {
//		return client.Dial("tcp", "192.168.0.68:10001")
//	})
//	if err != nil {
//		fmt.Println(err)
//	}
//	var a map[string]interface{}
//	err = sess.DB("xu_precise").C("taocat").Find(bson.M{"cid": "2"}).One(&a)
//	fmt.Println(err, a)

//	// Each ClientConn can support multiple interactive sessions,
//	// represented by a Session.
//	session, err := client.NewSession()
//	if err != nil {
//		panic("Failed to create session: " + err.Error())
//	}
//	defer session.Close()

//	// Once a Session is created, you can execute a single command on
//	// the remote side using the Run method.
//	var b bytes.Buffer
//	session.Stdout = &b
//	if err := session.Run("/bin/ls /"); err != nil {
//		panic("Failed to run: " + err.Error())
//	}
//	fmt.Println(b.String())

//}

//func getMongo(url string, f func() (net.Conn, error)) (*mgo.Session, error) {
//	info, err := mgo.ParseURL(url)
//	if err != nil {
//		return nil, err
//	}
//	info.Timeout = 10 * time.Second
//	info.DialServer = func(addr *mgo.ServerAddr) (net.Conn, error) {
//		return f()
//	}

//	return mgo.DialWithInfo(info)
//}

//func Dial(url string) (*Session, error) {
//	session, err := DialWithTimeout(url, 10*time.Second)
//	if err == nil {
//		session.SetSyncTimeout(1 * time.Minute)
//		session.SetSocketTimeout(1 * time.Minute)
//	}
//	return session, err
//}

//// DialWithTimeout works like Dial, but uses timeout as the amount of time to
//// wait for a server to respond when first connecting and also on follow up
//// operations in the session. If timeout is zero, the call may block
//// forever waiting for a connection to be made.
////
//// See SetSyncTimeout for customizing the timeout for the session.
//func DialWithTimeout(url string, timeout time.Duration) (*Session, error) {
//	info, err := ParseURL(url)
//	if err != nil {
//		return nil, err
//	}
//	info.Timeout = timeout
//	return DialWithInfo(info)
//}

package main

import (
	"fmt"
	"math"
	"net"
	"os"
	"time"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"code.google.com/p/go.crypto/ssh"
	"github.com/ngaut/log"
)

var (
	pemBytes = `-----BEGIN RSA PRIVATE KEY-----
MIIEowIBAAKCAQEAsI9auBcIIt6dYp+Na+k7MG1WVIQKWwyNBXHXkNY5gWScg2Zb
TEpao8ANolksIfCzk4NmOJHdBFU6Bx1Va12cPPGYEFmJFLbI0kb756K4zioku4e7
BYi/j9vfsRwA5XCyiPiTVDWzsqErlt+DonLxzVx4AALedW5wQto0ZdTRsnV1HgD4
i+L1433RgLwnHzHDVYOAO3rrDZHgUBfXc2mjdZaSugfZ7HMaGcN2hzFm11GKASO/
OGVUY/ColYU48oRWtxw5coaNGl/qwmFs3+U70LKEXF58IQQmso6ppSDVGrd+IGfI
OGhDoCao+RJ7oY9hXCRit085uzQ8MInwwbV6vQIDAQABAoIBAQCcGUIlvAcvfQ48
4b+RBpWUDTbkZhEZypDrnWju3tfctG1EJzzUyHA2klf7j0dbgoniA6xem2eCqy8w
lxisYgj+QMMmwWJW8/u9HZEdjFpDvDLZsfkBvZNPxx+QYKfSMr9GJi9rpkcHyULW
kyq4d1OdMwHNULwJquiJ0o288lmron9qyQ0eEyZ2V5QL1UCazfrXQTSwbKMRFNhq
HDTPQHN4KtgZh0/qnQfWWjqGk1o7g6QXp252jjYBJI7fHxTVNj7Imw0V8E0d/1u3
sJEQr81qKtwhSnRoXaRFzRMx1JuelqRFzKoteIBVVEbNKIML1HcR0zxFgQqHd8DI
Sp7eKBlpAoGBAOrevRlXDU6LjLiMQqSbgJp1oI1+fgG6W+DytP6tIDJJd/Ju4a3n
9QIMmFI+sWYZbyqZW/xaw3tcGAvLojrh/NLPG0+Rd4QXN6TI3d59JrzlRPz4LRe7
OOAKoMpsqBSvOYkx3a+D4JNxhCUNDs4smYOaKjpB5TXpoUty2u9W+v2nAoGBAMBx
rwXVkWPToAEAhZQUuwFILHwzApq1aRVGX5K/kM1FimceyLYoXT853OQuJDln6Awi
Ub3NcyLPDU3RpTDR1sltby0Djw+8S8L83JRu4axSQY65leYjorJCch0Lh8+znR54
SUpP8SkeQhv/z5zKNSGMqmrwzNUdrAjkBMaLk/j7AoGAYh3evWFCa9ecV9QwWvej
R+NvyOxY03v4ugZqWiIU2y0Z8KslmDLYhZyhXWpXTaG+cPtUFB4On9AfM35ELXkO
1zox3JGWbhYM5sgK99Esh6j3ov5CSDGsVtvZw/aUWN/Cl2+/fn0HKlE3tQq5bqPv
Fa0niuLQUC9jdFNs5qNdgu0CgYA4K7aTdF/ojGeigz16GIbw+9kIM3dqItNWQ2E4
GzQvxkF8ke6xxJxbDQ+dhp5KJzsC9612QhZ+LYNLmIqn8kfIKWoO8H/8btCKTHYx
2R+Dxcqe1yqarwIZF+3o7mmoxVtx/lgeGbFheBSByawWrqrNbRp89mZDOlLxkWSX
czwwqwKBgGqtq/v/Rys7TTAE9HkMjROvnT8+tUtSUoWcGfa4dQe9xmyDmfSGvLoN
NifztNrAQo7tV1QmHnA0ThEUNfdXwM6F1rntm2kKE+IgcSn7j7bGfNu0t9sDTBbE
AXuzTcd0cvQNXTbwqIg+3Yl7r+Y4ufUUz8Dw3oxBnUAEJOBbrP4S
-----END RSA PRIVATE KEY-----
`
)

var (
	mdbsession *mgo.Session
	mo_user    = "xu"
	mo_pwd     = "123456"
	mo_host    = "192.168.1.199"
	mo_port    = "27017"
	mo_db      = "xu_precise"
	mo_table   = "taocat"
)

//获取mongo数据库链接
func GetSession() *mgo.Session {
	var (
		mouser = mo_user
		mopwd  = mo_pwd
		mohost = mo_host
		moport = mo_port
		modb   = mo_db
	)

	if mdbsession == nil {
		var err error
		mdbsession, err = mgo.Dial(fmt.Sprintf("%s:%s@%s:%s/%s", mouser, mopwd, mohost, moport, modb))
		if err != nil {
			panic(err)
		}
	}
	//高并发下会关闭连接,ping下会恢复
	mdbsession.Ping()
	return mdbsession.Copy()
}
func main() {
	signer, err := ssh.ParsePrivateKey([]byte(pemBytes))
	if err != nil {
		panic(err)
	}

	//clientKey := &keychain{signer}

	config := &ssh.ClientConfig{
		User: "root",
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
	}
	c, err := ssh.Dial("tcp", "122.225.98.69:22", config)
	if err != nil {
		log.Error("unable to dial remote side:", err)
	}
	defer c.Close()

	sess, err := getMongo("192.168.0.86:27017/xu_precise", func() (net.Conn, error) {
		return c.Dial("tcp", "192.168.0.86:27017")
	})
	defer sess.Close()
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}

	nsess := GetSession()
	defer nsess.Close()

	count, _ := nsess.DB(mo_db).C("taocat").Count()
	pageSize := 1000
	pageCount := int(math.Ceil(float64(count) / float64(pageSize)))

	for i := 1; i <= pageCount; i++ {
		list := make([]map[string]interface{}, 0, pageSize)
		nsess.DB(mo_db).C("taocat").Find(bson.M{}).
			Limit(pageSize).Skip((i - 1) * pageSize).All(&list)
		var aaa = make([]interface{}, 0, len(list))
		for _, v := range list {
			aaa = append(aaa, bson.M(v))
			log.Info(v)
		}
		log.Warn(sess.DB(mo_db).C(mo_table).Insert(aaa...))
	}
	return
}

func getMongo(url string, f func() (net.Conn, error)) (*mgo.Session, error) {
	info, err := mgo.ParseURL(url)
	if err != nil {
		return nil, err
	}
	info.Timeout = 10 * time.Second
	info.DialServer = func(addr *mgo.ServerAddr) (net.Conn, error) {
		return f()
	}

	return mgo.DialWithInfo(info)
}

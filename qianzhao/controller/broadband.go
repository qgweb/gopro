// 宽带
package controller

import (
	"errors"
	"github.com/astaxie/beego/httplib"
	"github.com/labstack/echo"
	"github.com/qgweb/gopro/qianzhao/common/global"
	"github.com/qgweb/gopro/qianzhao/model"
	"log"
	"strings"
)

const (
	USER_QUERY_URL       = "http://js.vnet.cn/ProvinceForSPSearchUserName/services/ProvinceForSPSearchUserName?wsdl"
	USER_QUERY_PRODUCTID = "1100000300002204"
)

var (
	ErrProgram         = errors.New("程序发生异常")
	ErrData            = errors.New("数据格式出错")
	ErrUserUnup        = errors.New("用户未登录宽带")
	ErrNotJiangShuUser = errors.New("须为江苏电信校园宽带用户")
	ErrNotAllowUP      = errors.New("当前宽带环境不满足提速条件")
	ErrCode            = "3xx"
	SuccCode           = "200"
)

type BroadBand struct {
}

// 开启
func (this *BroadBand) Start(ctx *echo.Context) error {
	areacode, username, err := this.userQuery(ctx)
	if err != nil {
		return ctx.JSON(200, map[string]string{
			"code": ErrCode,
			"msg":  err.Error(),
		})
	}

	// 检测用户是否在白名单内
	baModel := model.BrandAccount{}
	if !baModel.AccountExist(model.BrandAccount{Account: username, Area: areacode}) {
		return ctx.JSON(200, map[string]string{
			"code": ErrCode,
			"msg":  ErrNotAllowUP.Error(),
		})
	}

	return ctx.JSON(200, map[string]string{
		"code": SuccCode,
		"msg":  "OK",
		"data": username,
	})
}

// 停止
func (this *BroadBand) Stop(ctx *echo.Context) error {
	return nil
}

// 用户宽带查询(username string, areacode string)
func (this *BroadBand) userQuery(ctx *echo.Context) (string, string, error) {
	var (
		ip = ctx.Form("ip")
	)

	req := httplib.Post(USER_QUERY_URL)
	req.Header("Content-Type", "text/xml; charset=utf-8")
	req.Header("SOAPAction", USER_QUERY_URL)
	req.Body(createUserSOAPXml("", "getUserProduct", ip, USER_QUERY_PRODUCTID))
	a := global.UserEnvelope{}
	err := req.ToXml(&a)
	if err != nil {
		log.Println("[BroadBand userQuery] 解析xml失败 ", err)
		return "", "", ErrProgram
	}

	res := a.By.GP.GetUserProductReturn
	resAry := strings.Split(res, "|")
	log.Println(res)
	if len(resAry) != 3 {
		return "", "", ErrProgram
	}

	if resAry[0] == "0" {
		return resAry[1], resAry[1], nil
	}
	if resAry[0] == "-999" {
		return "", "", ErrNotJiangShuUser
	}
	if resAry[0] == "-998" {
		return "", "", ErrProgram
	}

	return "", "", nil
}

func createUserSOAPXml(nameSpace string, methodName string, productid string, ip string) string {
	soapBody := "<?xml version=\"1.0\" encoding=\"utf-8\"?>"
	soapBody += "<soap12:Envelope xmlns:m=\"http://js.vnet.cn\" xmlns:xsi=\"http://www.w3.org/2001/XMLSchema-instance\" xmlns:xsd=\"http://www.w3.org/2001/XMLSchema\" xmlns:soap12=\"http://www.w3.org/2003/05/soap-envelope\">"
	soapBody += "<soap12:Body>"
	soapBody += "<" + methodName + " xmlns=\"" + nameSpace + "\">"
	//以下是具体参数
	soapBody += "<productid>" + productid + "</productid>"
	soapBody += "<ip>" + ip + "</ip>"
	soapBody += "</" + methodName + "></soap12:Body></soap12:Envelope>"
	return soapBody
}

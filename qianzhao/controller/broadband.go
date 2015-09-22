// 宽带
package controller

import (
	"bytes"
	"fmt"
	"log"
	"strings"
	"text/template"

	"github.com/astaxie/beego/httplib"
	"github.com/labstack/echo"
	"github.com/qgweb/gopro/qianzhao/common/global"
	"github.com/qgweb/gopro/qianzhao/model"
)

type ErrBrand struct {
	Code string
	Msg  string
}

type UserData struct {
	Area    string
	Account string
}

const (
	USER_QUERY_URL       = "http://js.vnet.cn/ProvinceForSPSearchUserName/services/ProvinceForSPSearchUserName?wsdl"
	USER_QUERY_PRODUCTID = "1100099900000000"
	DEFAULT_SPEED        = 1024 * 20 * 1.0 // 单位kb
)

var (
	ErrProgram         = &ErrBrand{"5000", "程序发生异常"}
	ErrNotJiangShuUser = &ErrBrand{"5001", "须为江苏电信校园宽带用户"}
	ErrNotAllowUP      = &ErrBrand{"5002", "当前宽带环境不满足提速条件"}
	ErrTimeOut         = &ErrBrand{"5003", "体验时间已结束"}
)

type BroadBand struct {
}

// 开启
func (this *BroadBand) Start(ctx *echo.Context) error {
	udata, err := this.userQuery(ctx)
	if err != nil {
		return ctx.JSON(200, map[string]string{
			"code": err.Code,
			"msg":  err.Msg,
		})
	}
	//udata := UserData{"0001", "10327158471"}

	// 检测用户是否在白名单内
	baModel := model.BrandAccount{}
	if !baModel.AccountExist(model.BrandAccount{Account: udata.Account, Area: udata.Area}) {
		return ctx.JSON(200, map[string]string{
			"code": ErrNotAllowUP.Code,
			"msg":  ErrNotAllowUP.Msg,
		})
	}

	// 检测用户时长
	ba, err1 := baModel.GetAccountInfo(udata.Account)
	if err1 != nil || ba.Id == "" {
		return ctx.JSON(200, map[string]string{
			"code": ErrProgram.Code,
			"msg":  ErrProgram.Msg,
		})
	}

	// 可以使用时长
	canTime := ba.TotalTime - ba.UsedTime
	if canTime == 0 {
		return ctx.JSON(200, map[string]string{
			"code": ErrTimeOut.Code,
			"msg":  ErrTimeOut.Msg,
		})
	}

	// 宽带提速比
	speed := 1 - float32(ba.DownBroadband)/float32(DEFAULT_SPEED)

	return ctx.JSON(200, map[string]string{
		"code": "200",
		"msg":  "OK",
		"data": fmt.Sprintf("%s|%d|%.2f", ba.Account, canTime, speed),
	})
}

// 停止
func (this *BroadBand) Stop(ctx *echo.Context) error {
	return nil
}

// 用户宽带查询(username string, areacode string)
func (this *BroadBand) userQuery(ctx *echo.Context) (*UserData, *ErrBrand) {
	var (
		ip = ctx.Request().RemoteAddr
		//ip = "121.237.226.1:11137"
	)

	req := httplib.Post(USER_QUERY_URL)
	req.Header("Content-Type", "text/xml; charset=utf-8")
	req.Header("SOAPAction", USER_QUERY_URL)
	req.Body(createUserSOAPXml(ip, USER_QUERY_PRODUCTID))
	a := global.UserEnvelope{}
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

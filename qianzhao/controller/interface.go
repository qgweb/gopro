package controller

import (
	"encoding/json"
	"github.com/labstack/echo"
	"github.com/ngaut/log"
	"github.com/qgweb/gopro/lib/encrypt"
	"github.com/qgweb/gopro/qianzhao/common/config"
	"github.com/qgweb/gopro/qianzhao/common/global"
	"github.com/qgweb/gopro/qianzhao/model"
	"io/ioutil"
)

type Interfacer struct{}

type JsonReturn struct {
	Code string `json:"code"`
	Msg  string `json:"msg"`
	Data string `json:"data"`
}

func (this *Interfacer) checkSafe(ctx *echo.Context) (bool, error) {
	key := ctx.Request().Header.Get("af0007298116140e4e8aa3e0cc763703")
	code := ctx.Request().Header.Get("bf97c2ebf72e9631502671a1b69bad3a")
	secret := config.GetInterface().Key("secret").String()

	if encrypt.DefaultMd5.Encode("qgweb_"+key+"_"+secret) == code {
		return true, nil
	}

	return false, ctx.JSON(200, JsonReturn{
		Code: global.CONTROLLER_INTERFACE_ERROR_CODE_HEADER,
		Msg:  global.CONTROLLER_INTERFACE_ERROR_MSG_HEADER,
		Data: "",
	})

}

// 账户名单
func (this *Interfacer) AccountList(ctx *echo.Context) error {
	// 验证是否有效
	if res, err := this.checkSafe(ctx); !res {
		return err
	}

	body, err := ioutil.ReadAll(ctx.Request().Body)
	if err != nil {
		log.Error("[interfacer AccountList] 读取body失败 ", err)
		return ctx.JSON(200, ctx.JSON(200, JsonReturn{
			Code: global.CONTROLLER_INTERFACE_ERROR_CODE_PROGRAM,
			Msg:  global.CONTROLLER_INTERFACE_ERROR_MSG_PROGRAM,
			Data: "",
		}))
	}

	accoutList := make([]model.BrandAccount, 0, 10)
	err = json.Unmarshal(body, &accoutList)
	if err != nil {
		log.Error("[interfacer AccountList] 转换白名单失败 ", err)
		return ctx.JSON(200, JsonReturn{
			Code: global.CONTROLLER_INTERFACE_ERROR_CODE_DATA,
			Msg:  global.CONTROLLER_INTERFACE_ERROR_MSG_DATA,
			Data: "",
		})
	}

	bmodel := model.BrandAccount{}

	for _, v := range accoutList {
		bmodel.Area = v.Area
		bmodel.Account = v.Account
		if bmodel.AccountExist(bmodel) {
			bmodel.EditAccount(v)
		} else {
			bmodel.AddBroadBand(v)
		}
	}

	ctx.HTML(200, "%v", accoutList)
	return nil
}

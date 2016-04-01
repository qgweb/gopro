package controller

import (
	"github.com/labstack/echo"
	"github.com/ngaut/log"
	"github.com/qgweb/gopro/lib/convert"
	"github.com/qgweb/gopro/qianzhao/model"
	"github.com/qgweb/new/lib/timestamp"
	"math/rand"
	"net/http"
	"time"
)

type Club struct {
	Base
}

var (
	awardChanNum = 30
	awardChan    = make(chan int, awardChanNum)
)

func (this *Club) Index(ctx *echo.Context) error {
	return ctx.Render(http.StatusOK, "club", nil)
}

func (this *Club) getRand(ary map[string]int) string {
	prosum := 0
	for _, v := range ary {
		prosum += v
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	for k, v := range ary {
		randNum := r.Intn(prosum)
		if randNum <= v {
			return k
		} else {
			prosum -= v
		}
	}

	return ""
}

func (this *Club) Winlist(ctx *echo.Context) error {
	//前3位在189/181/180/153/133/177中提取，中间4位用*隐去，
	//后四位由0-9随机组合生成。中奖纪录由1元话费、5元话费、10元话费按6:3：1概率生成。
	type Result struct {
		Phone  string
		Result string
		Date   string
	}
	var list = make(map[int]*Result)
	var rlist = make([]Result, 100)
	var numlist = []string{"189", "181", "180", "153", "133", "177"}
	for i := 0; i < 100; i++ {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		n := r.Intn(len(numlist))
		list[i] = &Result{}
		list[i].Phone = numlist[n] + "****" + convert.ToString(r.Intn(8999)+1000)
		list[i].Result = this.getRand(map[string]int{
			"1元话费充值卡":  600,
			"5元话费充值卡":  300,
			"10元话费充值卡": 100,
		})
		list[i].Date = time.Now().Format("2006-01-02")
	}

	for k, v := range list {
		rlist[k] = *v
	}

	return ctx.JSON(http.StatusOK, rlist)
}

func (this *Club) Turntable(ctx *echo.Context) error {
	// 验证是否登录
	if res, err := this.Base.IsLogin(ctx); !res {
		return err
	}

	if len(awardChan) > awardChanNum-1 {
		return ctx.JSON(http.StatusOK, map[string]interface{}{
			"code": 302,
			"msg":  "当前抽奖人数过多，请稍后再试^_^",
		})
	}
	awardChan <- 1
	defer func() {
		<-awardChan
	}()

	ui := this.Base.GetUserInfo(ctx)
	aw := model.Award{}
	resMsg := []string{
		"谢谢参与！",
		"恭喜您抽中1元话费充值卡！",
		"恭喜您抽中5元话费充值卡！",
		"恭喜您抽中10元话费充值卡！",
	}
	v, err := aw.HaveJoin(ui.Id, timestamp.GetDayTimestamp(0))
	if err != nil {
		log.Error(err)
	}
	if v {
		return ctx.JSON(http.StatusOK, map[string]interface{}{
			"code": 301,
			"msg":  "您今天抽奖次数已经用完，请明天再来！",
		})
	}
	n, code, err := aw.Get(ui.Id, map[int]int{
		0: 608,
		1: 333,
		2: 53,
		3: 6,
	})

	if err != nil {
		log.Error(err)
	}

	return ctx.JSON(http.StatusOK, map[string]interface{}{
		"code":  "200",
		"res":   resMsg[n],
		"num":   n,
		"rcode": code,
	})
}

func (this *Club) Mylist(ctx *echo.Context) error {
	//验证是否登录
	if res, err := this.Base.IsLogin(ctx); !res {
		return err
	}
	var aw = model.Award{}
	list, err := aw.MyRecord(this.Base.GetUserInfo(ctx).Id)
	if err != nil {
		log.Error(err)
	}
	return ctx.JSON(http.StatusOK, list)
}

package controller

import (
	"fmt"
	"github.com/gobuild/log"
	"github.com/labstack/echo"
	"github.com/qgweb/gopro/qianzhao/common/config"
	"github.com/qgweb/gopro/qianzhao/model"
	"github.com/qgweb/new/lib/convert"
	"github.com/qgweb/new/lib/timestamp"
	"net/http"
	"strings"
	"time"
	"math/rand"
)

type Club2 struct {
	Base
}

const CAN_AWARD = "can_award"

var (
	awardChanNum = 50
	awardChan = make(chan int, awardChanNum)
	bt, _ = config.GetDefault().Key("beginTime").Int64()
	BTime = time.Unix(bt, 0)
)

func (this *Club2) setCanAward(ctx echo.Context) {
	if v, _ := this.Base.GetSess(ctx, CAN_AWARD); v == nil {
		var um model.User
		ui := this.Base.GetUserInfo(ctx)
		if ui.Phone != "" {
			if res, err := um.CanAward(ui.Phone); err == nil && res {
				this.Base.SetSess(ctx, "can_award", true)
			} else {
				log.Error(err)
			}
		}
	}
}

func (this *Club2) canAward(ctx echo.Context) bool {
	v, _ := this.Base.GetSess(ctx, CAN_AWARD)
	return v != nil && v.(bool)
}

func (this *Club2) construct(ctx echo.Context) {
	this.setCanAward(ctx)
}

func (this *Club2) PrevIndex(ctx echo.Context) error {
	return ctx.HTML(200, `
<html>
<head>
<title>惊喜开学季,话费送不停</title>
<style>
body {margin-left: 0px;margin-top: 0px;margin-right: 0px;margin-bottom:  0px;overflow: hidden;}
</style>
</head>
<body>
<iframe src='http://qianzhao.221su.com/club2' width='100%' height='100%'  frameborder='0' name="_blank" id="_blank" ></iframe>
</body>
</html>
	`)
}

func (this *Club2) formatAwardInfo(name string, at int, ist bool) string {
	var resMsg = []string{
		"谢谢参与！",
		"抽中5元话费充值卡！",
		"抽中10元话费充值卡！",
		"抽中20元话费充值卡！",
	}
	var nameAry = []rune(name)

	for i := 3; i < len(nameAry); i++ {
		nameAry[i] = []rune("*")[0]
	}
	//nameAry[0] = []rune("*")[0]
	//nameAry[len(nameAry) - 1] = []rune("*")[0]

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	na := 0
	n := r.Int63n(1000)
	switch  {
	case n > 0 && n < 800:
		na = 1
	case n >= 800 && n < 900:
		na = 2
	case n >= 900 && n < 999:
		na = 3
	}
	log.Error(na)
	if ist {
		na = at
	}
	return "恭喜" + string(nameAry) + resMsg[na]
}

func (this *Club2) Index(ctx echo.Context) error {
	var sm model.Sign
	var um model.User
	var am model.Award
	var uid = convert.ToInt(this.GetUserInfo(ctx).Id)
	var info = make(map[string]interface{})

	info["lottery_count"] = 0
	info["list1"] = map[string]string{}
	info["list2"] = map[string]string{}
	info["is_sign"] = false

	sm.Reset(uid)
	this.construct(ctx)
	if uid != 0 {
		s, err := sm.GetInfo(uid)
		if err != nil {
			log.Error(err)
		} else {
			info["sign"] = s.History
			info["is_sign"] = sm.HasSign(uid)
		}
		if n, err := um.GetAwardCount(convert.ToString(uid)); err == nil {
			info["lottery_count"] = n
		} else {
			log.Error(err)
		}
	}
	if l, err := am.Records("0"); err == nil {
		if len(l) % 2 > 0 {
			l = append(l[:], l[:]...)
		}
		for k, v := range l {
			l[k]["tag"] = "0"
			if k % 2 != 0 {
				l[k]["tag"] = "1"
			}
			l[k]["title"] = this.formatAwardInfo(v["username"], convert.ToInt(v["awards_type"]),false)
		}
		info["list1"] = l
	}
	if l, err := am.Records("1"); err == nil {
		if len(l) % 2 > 0 {
			l = append(l[:], l[:]...)
		}
		for k, v := range l {
			l[k]["tag"] = "0"
			if k % 2 != 0 {
				l[k]["tag"] = "1"
			}
			l[k]["title"] = this.formatAwardInfo(v["username"], convert.ToInt(v["awards_type"]),true)
		}
		info["list2"] = l
	}

	//info["sign"] = "11000"
	return ctx.Render(200, "club2", info)
}

// 签到
func (this *Club2) Sign(ctx echo.Context) error {
	if res, err := this.Base.IsLogin(ctx); !res {
		return err
	}

	this.construct(ctx)
	if !this.canAward(ctx) {
		return ctx.JSON(http.StatusOK, map[string]interface{}{
			"ret": -1,
			"msg": "您非校园用户，无法参与活动",
		})
	}
	var sm model.Sign
	var um model.User
	var uid = convert.ToInt(this.GetUserInfo(ctx).Id)

	if sm.HasSign(uid) {
		return ctx.JSON(http.StatusOK, map[string]interface{}{
			"ret": -1,
			"msg": "您已经签过了，请明天再试",
		})
	}
	r, err := sm.Add(uid, func() {
		//签到5次
		um.IncrAwardCount(uid, 1)
	})
	fmt.Println(r, err)
	if err == nil {
		return ctx.JSON(http.StatusOK, map[string]interface{}{
			"ret":  0,
			"msg":  "签到成功",
			"data": r,
		})
	}
	if err != nil {
		log.Error(err)
	}
	return ctx.JSON(http.StatusOK, map[string]interface{}{
		"ret": -1,
		"msg": "签到失败，请重试",
	})
}

// 猜字谜
func (this *Club2) Gword(ctx echo.Context) error {
	if res, err := this.Base.IsLogin(ctx); !res {
		return err
	}

	this.construct(ctx)
	if !this.canAward(ctx) {
		return ctx.JSON(http.StatusOK, map[string]interface{}{
			"ret": -1,
			"msg": "您非校园用户，无法参与活动",
		})
	}
	var wm model.Word
	var am model.Award
	var uid = convert.ToInt(this.GetUserInfo(ctx).Id)
	var w = strings.TrimSpace(ctx.FormValue("w"))
	//判断是否猜过
	if r, _ := wm.Has(uid); r {
		return ctx.JSON(http.StatusOK, map[string]interface{}{
			"ret": -1,
			"msg": "您已经猜过了！",
		})
	}
	wi, err := wm.Get()
	if err != nil {
		log.Error(err)
		return ctx.JSON(http.StatusOK, map[string]interface{}{
			"ret": -1,
			"msg": "系统发生错误！",
		})
	}

	if wi.Time != convert.ToInt(timestamp.GetHourTimestamp(0)) {
		return ctx.JSON(http.StatusOK, map[string]interface{}{
			"ret": -1,
			"msg": "字谜活动还没开始，敬请期待！",
		})
	}

	if wi.Word == w {
		log.Info(wi.HasCount())
		if n, err := wi.HasCount(); err == nil && n < 20 {
			n, code, err := am.Word(this.GetUserInfo(ctx).Id, true)
			if err != nil {
				log.Error(err)
				return ctx.JSON(http.StatusOK, map[string]interface{}{
					"ret": -1,
					"msg": "系统发生错误",
				})
			}
			return ctx.JSON(http.StatusOK, map[string]interface{}{
				"ret": 0,
				"msg": "恭喜您猜中了",
				"data": map[string]interface{}{
					"n": n,
					"c": code,
				},
			})
		} else {
			am.Word(this.GetUserInfo(ctx).Id, false)
			return ctx.JSON(http.StatusOK, map[string]interface{}{
				"ret": -1,
				"msg": " 本轮话费已抢完，请明日继续",
			})
		}
	}

	am.Word(this.GetUserInfo(ctx).Id, false)
	return ctx.JSON(http.StatusOK, map[string]interface{}{
		"ret": -1,
		"msg": "非常遗憾，您没有猜中！",
	})
	return nil
}

// 转盘
func (this *Club2) Turntable(ctx echo.Context) error {
	// 验证是否登录
	if res, err := this.Base.IsLogin(ctx); !res {
		return err
	}

	this.construct(ctx)
	if !this.canAward(ctx) {
		return ctx.JSON(http.StatusOK, map[string]interface{}{
			"ret": -1,
			"msg": "您非校园用户，无法参与活动",
		})
	}

	var uid = convert.ToInt(this.GetUserInfo(ctx).Id)
	var um model.User
	if c, _ := um.GetAwardCount(this.GetUserInfo(ctx).Id); c == 0 {
		return ctx.JSON(http.StatusOK, map[string]interface{}{
			"ret": -1,
			"msg": "您目前没有抽奖机会，请参与签到获取",
		})
	}

	if len(awardChan) > awardChanNum - 1 {
		return ctx.JSON(http.StatusOK, map[string]interface{}{
			"code": 302,
			"msg":  "当前抽奖人数过多，请稍后再试^_^",
		})
	}
	awardChan <- 1
	defer func() {
		<-awardChan
	}()

	var acm model.AwardCount
	ui := this.Base.GetUserInfo(ctx)
	aw := model.Award{}
	resMsg := []string{
		"谢谢参与！",
		"恭喜您抽中5元话费充值卡！",
		"恭喜您抽中10元话费充值卡！",
		"恭喜您抽中20元话费充值卡！",
	}

	ar, err := acm.Get()
	if err != nil {
		return ctx.JSON(http.StatusOK, map[string]interface{}{
			"code": 301,
			"msg":  "系统发生错误！",
		})
	}
	fmt.Println(um.IncrAwardCount(uid, -1))
	n, code, err := aw.Get(ui.Id, map[int]int{
		0: 1000 - ar.Five - ar.Ten - ar.Twenty,
		1: ar.Five,
		2: ar.Ten,
		3: ar.Twenty,
	}, 0)

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

func (this *Club2) Mylist(ctx echo.Context) error {
	//验证是否登录
	if res, err := this.Base.IsLogin(ctx); !res {
		return err
	}
	var aw = model.Award{}
	list, err := aw.MyRecord(this.Base.GetUserInfo(ctx).Id)
	if err != nil {
		log.Error(err)
	}
	resMsg := []string{
		"谢谢参与！",
		"恭喜您抽中5元话费充值卡！",
		"恭喜您抽中10元话费充值卡！",
		"恭喜您抽中20元话费充值卡！",
	}
	sourceMsg := []string{
		"转盘",
		"猜字谜",
	}
	for k, v := range list {
		list[k]["title"] = resMsg[convert.ToInt(v["awards_type"])]
		list[k]["source"] = sourceMsg[convert.ToInt(v["source"])]
	}
	return ctx.JSON(http.StatusOK, list)
}

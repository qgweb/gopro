package router

import (
	"github.com/labstack/echo"
	"github.com/qgweb/gopro/qianzhao/controller"
)

func Router(e *echo.Echo) {
	//// 秒速浏览器用户注册
	e.Post("/Api/v0.1/user/register", (&controller.User{}).Register)

	//// 秒速浏览器用户登录
	e.Post("/Api/v0.1/user/login", (&controller.User{}).Login)
	e.Get("/Api/v0.1/user/is_login", (&controller.User{}).IsLogin)
	e.Get("/Api/v0.1/user/logout", (&controller.User{}).LoginOut)

	//// 用户中心
	e.Get("/user", (&controller.User{}).MemberCenter)
	e.Get("/user/edit", (&controller.User{}).Edit)
	e.Post("/user/editname", (&controller.User{}).EditUsername)
	e.Post("/user/editemail", (&controller.User{}).EditUseremail)
	e.Post("/user/editpic", (&controller.User{}).EditUserpic)
	e.Post("/user/getcode", (&controller.User{}).GetUserPhoneCode)
	e.Post("/user/editphone", (&controller.User{}).EditUserphone)
	e.Post("/user/uploadpic", (&controller.User{}).UploadPic)
	e.Get("/ip", (&controller.User{}).GetIp)

	//// 用户收藏夹
	e.Post("/Api/v0.1/backup_favorite", (&controller.Favorite{}).BackupFavorite)
	e.Post("/Api/v0.1/get_favorite", (&controller.Favorite{}).GetFavorite)

	//// 宽带账号
	e.Post("/Api/v0.1/verify_Bandwith", (&controller.User{}).VerifyBandwith)
	e.Get("/Api/v0.1/get_Bandwith", (&controller.User{}).GetBandwith)

	//// 123
	e.Post("/Api/v0.1/speedup_open", (&controller.Ebit{}).SpeedupOpen)
	e.Get("/Api/v0.1/speedup_open", (&controller.Ebit{}).SpeedupOpen)

	e.Get("/Api/v0.1/speedup_open_check", (&controller.Ebit{}).SpeedupOpenCheck)
	e.Post("/Api/v0.1/speedup_open_check", (&controller.Ebit{}).SpeedupOpenCheck)

	e.Get("/Api/v0.1/speedup_check", (&controller.Ebit{}).SpeedupCheck)
	e.Post("/Api/v0.1/speedup_check", (&controller.Ebit{}).SpeedupCheck)

	//// 测速
	e.Get("/operate/speedup_prepare", (&controller.Operate{}).SpeedupPrepare)
	e.Get("/Api/v0.1/user/speed_test", (&controller.User{}).SpeedTest)
	e.Post("/operate/speedup_open_check", (&controller.Ebit{}).SpeedupOpenCheck)
	e.Post("/operate/speedup_open", (&controller.Ebit{}).SpeedupOpen)

	//// 接口
	e.Post("/interface/account", (&controller.Interfacer{}).AccountList)

	//// 宽带（双速网）
	e.Post("/broadbrand/start", (&controller.BroadBand{}).Start)
	e.Get("/broadbrand/resettime", (&controller.BroadBand{}).ResetTime)
	e.Get("/version/update", (&controller.Index{}).Update)

	//// 统计
	e.Get("/app", (&controller.Statistics{}).Download)
	e.Post("/stats/day", (&controller.Statistics{}).DayActivity)
	e.Post("/stats/sidbar", (&controller.Statistics{}).SideBar)

	//// 反馈
	e.Get("/feedback", (&controller.FeedBack{}).Index)
	e.Post("/feedback/post", (&controller.FeedBack{}).Post)
	e.Get("/feedback/pic", (&controller.FeedBack{}).Pic)

	/////手机验证码
	e.Get("/getcode", ((&controller.User{}).GetPhoneCode))

	//// 转盘
	//e.Get("/club", ((&controller.Club{}).Index))
	//e.Get("/club/turntable", ((&controller.Club{}).Turntable))
	//e.Get("/club/winlist", (&controller.Club{}).Winlist)
	//e.Get("/club/mywin", ((&controller.Club{}).Mylist))

	//// 首页
	e.Get("/index/getmainpage", (&controller.Index{}).MainPage)

	////找回密码
	e.Get("/user/pwd", (&controller.User{}).FindPwdView)
	e.GET("/forget/code", (&controller.User{}).GetPhoneCodeByPwd)
	e.Post("/forget/pwd", (&controller.User{}).FindPwd)
	e.Get("/user/pwdcode", (&controller.User{}).GetFindPwdCode)

	////首页
	e.Get("/index", (&controller.Index{}).Index)

	//// 新版活动
	e.Get("/club", (&controller.Club2{}).Index)
	e.Get("/club/sign",(&controller.Club2{}).Sign)
	e.Post("/club/word",(&controller.Club2{}).Gword)
	e.Get("/club/tun",(&controller.Club2{}).Turntable)
	e.Post("/club/mrecord",(&controller.Club2{}).Mylist)
}
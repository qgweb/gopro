package controller

import (
	"log"

	"github.com/goweb/gopro/lib/encrypt"
	"github.com/goweb/gopro/qianzhao/common/function"
	"github.com/goweb/gopro/qianzhao/common/global"
	"github.com/goweb/gopro/qianzhao/common/session"
	"github.com/goweb/gopro/qianzhao/model"

	"net/http"

	"github.com/labstack/echo"
)

//

type User struct {
	Base
}

// 登录
func (this *User) Login(ctx *echo.Context) error {
	sess, err := session.GetSession(ctx)
	if err != nil {
		log.Println("获取session失败：", err)
	}

	defer sess.SessionRelease(ctx.Response())
	// return ctx.String(200, "fff")

	var (
		userName = ctx.Form("username")
		password = ctx.Form("password")
		umodel   = model.User{}
	)

	if !umodel.UserNameExist(userName) {
		return ctx.JSON(http.StatusOK, map[string]string{
			"code": global.CONTROLLER_CODE_ERROR,
			"msg":  global.CONTROLLER_USER_NOT_EXIST,
		})
	}

	if umodel.UserExist(userName, password) {
		sid := encrypt.DefaultMd5.Encode(userName + function.GetTimeUnix())
		umodel.Update(map[string]interface{}{"sid": sid}, map[string]interface{}{"username": userName})
		uinfo := umodel.UserInfo(userName)
		avatar := uinfo.Avatar
		if avatar == "" {
			avatar = "/upload/avatar.png"
		}

		sess.Set(global.SESSION_USER_INFO, uinfo) //存入session

		return ctx.JSON(http.StatusOK, map[string]string{
			"code":   global.CONTROLLER_CODE_SUCCESS,
			"msg":    global.CONTROLLER_USER_LOGIN_SUCCESS,
			"avatar": avatar,
			"sid":    sid,
		})
	} else {
		return ctx.JSON(http.StatusOK, map[string]string{
			"code": global.CONTROLLER_CODE_ERROR,
			"msg":  global.CONTROLLER_USER_CHECK_ERROR,
		})
	}
}

// 注册
func (this *User) Register(ctx *echo.Context) error {
	var (
		username = ctx.Form("username")
		password = ctx.Form("password")
		pwd      = ctx.Form("pwd")
		app_uid  = ctx.Form("app_uid")
		umodel   = model.User{}
	)

	if pwd != password {
		return ctx.JSON(http.StatusOK, map[string]string{
			"code": global.CONTROLLER_CODE_ERROR,
			"msg":  global.CONTROLLER_USER_USERPWD_ERROR,
		})
	}

	// 验证用户名是否存在
	if umodel.UserNameExist(username) {
		return ctx.JSON(http.StatusOK, map[string]string{
			"code": global.CONTROLLER_CODE_ERROR,
			"msg":  global.CONTROLLER_USER_USERNAME_EXIST_ERROR,
		})
	}

	// 存在app_uid 走关联,不存在,则走重新注册
	if app_uid != "" {
		bpasswd := function.GetBcrypt([]byte(password))
		res := umodel.Update(map[string]interface{}{
			"username": username,
			"password": bpasswd,
		}, map[string]interface{}{
			"app_uid": app_uid,
		})

		if res {
			return ctx.JSON(http.StatusOK, map[string]string{
				"code": global.CONTROLLER_CODE_SUCCESS,
				"msg":  global.CONTROLLER_USER_REGISTER_REF_SUCCESS,
			})
		} else {
			return ctx.JSON(http.StatusOK, map[string]string{
				"code": global.CONTROLLER_CODE_ERROR,
				"msg":  global.CONTROLLER_USER_REGISTER_REF_ERROR,
			})
		}
	} else {
		res := umodel.UserRegister(username, password)
		if res {
			return ctx.JSON(http.StatusOK, map[string]string{
				"code": global.CONTROLLER_CODE_SUCCESS,
				"msg":  global.CONTROLLER_USER_REGISTER_SUCCESS,
			})
		} else {
			return ctx.JSON(http.StatusOK, map[string]string{
				"code": global.CONTROLLER_CODE_ERROR,
				"msg":  global.CONTROLLER_USER_REGISTER_ERROR,
			})
		}
	}

	return nil
}

// 编辑
func (this *User) Edit(ctx *echo.Context) error {
	return ctx.String(http.StatusOK, "暂无")
}

// 是否登录
func (this *User) IsLogin(ctx *echo.Context) error {
	return this.Base.IsLogin(ctx)
}

// 退出登录
func (this *User) LoginOut(ctx *echo.Context) error {
	sess, err := session.GetSession(ctx)
	if err != nil {
		log.Println("获取session失败：", err)
	}

	defer sess.SessionRelease(ctx.Response())

	sess.Delete(global.SESSION_USER_INFO)

	return ctx.JSON(http.StatusOK, map[string]string{
		"code": global.CONTROLLER_CODE_SUCCESS,
		"msg":  global.CONTROLLER_USER_LOGINOUT,
	})
}

// 用户
func (this *User) MemberCenter(ctx *echo.Context) error {
	sess, err := session.GetSession(ctx)
	if err != nil {
		log.Println("获取session失败：", err)
	}

	defer sess.SessionRelease(ctx.Response())

	if u, ok := sess.Get(global.SESSION_USER_INFO).(model.User); ok {
		return ctx.Render(200, "usercenter", struct{ Username string }{u.Name})
	} else {
		return ctx.JSON(http.StatusOK, map[string]string{
			"code": global.CONTROLLER_CODE_ERROR,
			"msg":  global.CONTROLLER_USER_NEEDLOGIN,
		})
	}
}

// 速度测试
func (this *User) SpeedTest(ctx *echo.Context) error {
	return ctx.Render(200, "speedtest", "")
}

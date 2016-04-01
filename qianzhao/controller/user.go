package controller

import (
	"encoding/json"
	"github.com/ngaut/log"
	"github.com/qgweb/gopro/qianzhao/common/Sms"
	"net/url"

	"github.com/qgweb/gopro/lib/convert"
	"github.com/qgweb/gopro/lib/grab"

	"github.com/qgweb/gopro/lib/encrypt"
	"github.com/qgweb/gopro/qianzhao/common/function"
	"github.com/qgweb/gopro/qianzhao/common/global"
	"github.com/qgweb/gopro/qianzhao/common/redis"
	"github.com/qgweb/gopro/qianzhao/common/session"
	"github.com/qgweb/gopro/qianzhao/model"

	"net/http"

	oredis "github.com/garyburd/redigo/redis"
	"github.com/labstack/echo"
	"strings"
)

type User struct {
	Base
}

func (this *User) getIp(ctx *echo.Context) string {
	return strings.Split(ctx.Request().RemoteAddr, ":")[0]
}

// 登录
func (this *User) Login(ctx *echo.Context) error {
	sess, err := session.GetSession(ctx)
	if err != nil {
		log.Error("获取session失败：", err)
		return err
	}

	defer sess.SessionRelease(ctx.Response())
	// return ctx.String(200, "fff")

	var (
		userName, _ = url.QueryUnescape(ctx.Form("username"))
		password, _ = url.QueryUnescape(ctx.Form("password"))
		umodel      = model.User{}
	)

	if !umodel.UserNameExist(userName) {
		return ctx.JSON(http.StatusOK, map[string]string{
			"code": global.CONTROLLER_CODE_ERROR,
			"msg":  global.CONTROLLER_USER_NOT_EXIST,
		})
	}

	if umodel.UserExist(userName, password) {
		uinfo := umodel.UserInfo(userName)
		sid := encrypt.DefaultMd5.Encode(userName + function.GetTimeUnix())
		umodel.Update(map[string]interface{}{"sid": sid}, map[string]interface{}{"id": uinfo.Id})
		avatar := uinfo.Avatar
		if avatar == "" {
			avatar = "/upload/avatar.png"
		}

		sess.Set(global.SESSION_USER_INFO, uinfo) //存入session

		return ctx.JSON(http.StatusOK, map[string]string{
			"code":     global.CONTROLLER_CODE_SUCCESS,
			"msg":      global.CONTROLLER_USER_LOGIN_SUCCESS,
			"avatar":   avatar,
			"sid":      sid,
			"username": userName,
			"phone":    uinfo.Phone,
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
	sess, err := session.GetSession(ctx)
	if err != nil {
		log.Error("获取session失败：", err)
		return err
	}
	defer sess.SessionRelease(ctx.Response())

	var (
		phone, _    = url.QueryUnescape(ctx.Form("phone"))
		password, _ = url.QueryUnescape(ctx.Form("password"))
		email, _    = url.QueryUnescape(ctx.Form("email"))
		pwd, _      = url.QueryUnescape(ctx.Form("pwd"))
		app_uid, _  = url.QueryUnescape(ctx.Form("app_uid"))
		code, _     = url.QueryUnescape(ctx.Form("code"))
		codekey     = "REG_CODE_" + phone
		umodel      = model.User{}
		rdb         = redis.Get()
		rkey        = "REGISTER_" + this.getIp(ctx)
		vcount      = 5
	)

	defer rdb.Close()

	vnum, err := oredis.Int(rdb.Do("GET", rkey))
	if err != nil && err != oredis.ErrNil {
		return ctx.JSON(http.StatusOK, map[string]string{
			"code": global.CONTROLLER_CODE_ERROR,
			"msg":  "系统发生错误,稍后重试",
		})
	}

	if vnum >= vcount {
		return ctx.JSON(http.StatusOK, map[string]string{
			"code": global.CONTROLLER_CODE_ERROR,
			"msg":  "注册访问次数过多,请明天再试",
		})
	}

	if phone == "" {
		return ctx.JSON(http.StatusOK, map[string]string{
			"code": global.CONTROLLER_CODE_ERROR,
			"msg":  "手机号码不能为空",
		})
	}

	if email == "" {
		return ctx.JSON(http.StatusOK, map[string]string{
			"code": global.CONTROLLER_CODE_ERROR,
			"msg":  "邮箱不能为空",
		})
	}

	if pwd != password {
		return ctx.JSON(http.StatusOK, map[string]string{
			"code": global.CONTROLLER_CODE_ERROR,
			"msg":  "两次密码不一致",
		})
	}

	// 验证手机号码是否存在
	if umodel.PhoneExist(phone) {
		return ctx.JSON(http.StatusOK, map[string]string{
			"code": global.CONTROLLER_CODE_ERROR,
			"msg":  global.CONTROLLER_USER_USERNAME_EXIST_ERROR,
		})
	}

	// 验证邮箱是否存在
	if umodel.EmailExist(email) {
		return ctx.JSON(http.StatusOK, map[string]string{
			"code": global.CONTROLLER_CODE_ERROR,
			"msg":  global.CONTROLLER_USER_EMAIL_EXIST_ERROR,
		})
	}

	if code == "" {
		return ctx.JSON(http.StatusOK, map[string]string{
			"code": global.CONTROLLER_CODE_ERROR,
			"msg":  "验证码不能为空",
		})
	}

	if convert.ToString(sess.Get(codekey)) != code {
		return ctx.JSON(http.StatusOK, map[string]string{
			"code": global.CONTROLLER_CODE_ERROR,
			"msg":  "验证码错误",
		})
	}

	sess.Delete(codekey)

	// 存在app_uid 走关联,不存在,则走重新注册
	if app_uid != "" {
		bpasswd := function.GetBcrypt([]byte(password))
		res := umodel.Update(map[string]interface{}{
			"phone":    phone,
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
		res := umodel.UserRegister(phone, email, password)
		if res {
			// 记录访问次数
			key := "REGISTER_" + this.getIp(ctx)
			rdb.Do("INCR", key)
			if vnum == 0 {
				rdb.Do("EXPIRE", key, 86400)
			}

			if u := umodel.UserInfo(phone); u.Name != "" {
				sess.Set(global.SESSION_USER_INFO, u) //存入session
			}

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

// 修改用户昵称
func (this *User) EditUsername(ctx *echo.Context) error {
	res, _ := this.Base.IsLogin(ctx)
	if !res {
		return nil
	}

	var username = ctx.Form("username")
	var umodel = model.User{}

	if username == "" {
		return ctx.JSON(200, map[string]string{
			"code": "300",
			"msg":  "用户昵称不能为空",
		})
	}

	if un := umodel.UserInfo(username); un.Name != "" &&
		un.Name != umodel.UserInfoById(this.Base.GetUserInfo(ctx).Id).Name {
		return ctx.JSON(200, map[string]string{
			"code": "300",
			"msg":  "用户昵称已存在",
		})
	} else if un.Name == username {
		return ctx.JSON(200, map[string]string{
			"code": "300",
			"msg":  "用户昵称没有改变",
		})
	}

	if umodel.Update(map[string]interface{}{"username": username},
		map[string]interface{}{"id": this.Base.GetUserInfo(ctx).Id}) {
		return ctx.JSON(200, map[string]string{
			"code": "200",
			"msg":  "修改成功",
		})
	} else {
		return ctx.JSON(200, map[string]string{
			"code": "300",
			"msg":  "修改失败",
		})
	}
}

// 修改用户头像
func (this *User) EditUserpic(ctx *echo.Context) error {
	res, _ := this.Base.IsLogin(ctx)
	if !res {
		return nil
	}

	var pic = ctx.Form("pic")
	var umodel = model.User{}

	if pic == "" {
		return ctx.JSON(200, map[string]string{
			"code": "300",
			"msg":  "用户头像不能为空",
		})
	}

	log.Error(function.ThumbPic(function.GetBasePath()+"/public/"+pic, 150, 150))

	if umodel.Update(map[string]interface{}{"avatar": pic},
		map[string]interface{}{"id": this.Base.GetUserInfo(ctx).Id}) {
		return ctx.JSON(200, map[string]string{
			"code": "200",
			"msg":  "修改成功",
		})
	} else {
		return ctx.JSON(200, map[string]string{
			"code": "300",
			"msg":  "修改失败",
		})
	}
}

// 上传头像
func (this *User) UploadPic(ctx *echo.Context) error {
	res, _ := this.Base.IsLogin(ctx)
	if !res {
		return nil
	}
	var cb = ctx.Form("callback")
	path, err := UploadPic(ctx, "photo", false)

	if err != nil {
		ebs, _ := json.Marshal(map[string]string{
			"code": "300",
			"msg":  err.Error(),
		})
		return ctx.HTML(200, "<script>parent.%s(%s)</script>", cb, ebs)
	}

	ebs, _ := json.Marshal(map[string]string{
		"code": "200",
		"msg":  path,
	})
	return ctx.HTML(200, "<script>parent.%s(%s)</script>", cb, ebs)

}

// 修改用户email
func (this *User) EditUseremail(ctx *echo.Context) error {
	res, _ := this.Base.IsLogin(ctx)
	if !res {
		return nil
	}

	var email = ctx.Form("email")
	var umodel = model.User{}

	if email == "" {
		return ctx.JSON(200, map[string]string{
			"code": "300",
			"msg":  "用户邮箱不能为空",
		})
	}

	if un := umodel.UserInfo(email); un.Email != "" &&
		un.Email != umodel.UserInfoById(this.Base.GetUserInfo(ctx).Id).Email {
		return ctx.JSON(200, map[string]string{
			"code": "300",
			"msg":  "用户邮箱已存在",
		})
	} else if un.Email == email {
		return ctx.JSON(200, map[string]string{
			"code": "300",
			"msg":  "用户邮箱没有改变",
		})
	}

	if umodel.Update(map[string]interface{}{"email": email},
		map[string]interface{}{"id": this.Base.GetUserInfo(ctx).Id}) {
		return ctx.JSON(200, map[string]string{
			"code": "200",
			"msg":  "修改成功",
		})
	} else {
		return ctx.JSON(200, map[string]string{
			"code": "300",
			"msg":  "修改失败",
		})
	}
}

// 修改用户手机
func (this *User) EditUserphone(ctx *echo.Context) error {
	res, _ := this.Base.IsLogin(ctx)
	if !res {
		return nil
	}

	sess, err := session.GetSession(ctx)
	if err != nil {
		log.Error("获取session失败：", err)
		return err
	}
	defer sess.SessionRelease(ctx.Response())

	var phone = ctx.Form("phone")
	var code = ctx.Form("code")
	var umodel = model.User{}

	if phone == "" {
		return ctx.JSON(200, map[string]string{
			"code": "300",
			"msg":  "用户手机号码不能为空",
		})
	}

	if code == "" {
		return ctx.JSON(200, map[string]string{
			"code": "300",
			"msg":  "验证码不能为空",
		})
	}

	if code != sess.Get("USER_CODE_"+phone).(string) {
		return ctx.JSON(200, map[string]string{
			"code": "300",
			"msg":  "验证码错误",
		})
	}

	if un := umodel.UserInfo(phone); un.Phone != "" &&
		un.Phone != umodel.UserInfoById(this.Base.GetUserInfo(ctx).Id).Phone {
		return ctx.JSON(200, map[string]string{
			"code": "300",
			"msg":  "用户手机号码已存在",
		})
	} else if un.Email == phone {
		return ctx.JSON(200, map[string]string{
			"code": "300",
			"msg":  "用户手机号码没有改变",
		})
	}

	if umodel.Update(map[string]interface{}{"phone": phone},
		map[string]interface{}{"id": this.Base.GetUserInfo(ctx).Id}) {
		return ctx.JSON(200, map[string]string{
			"code": "200",
			"msg":  "修改成功",
		})
	} else {
		return ctx.JSON(200, map[string]string{
			"code": "300",
			"msg":  "修改失败",
		})
	}
}

// 是否登录
func (this *User) IsLogin(ctx *echo.Context) error {
	sess, err := session.GetSession(ctx)
	if err != nil {
		log.Error("获取session失败：", err)
		return err
	}

	defer sess.SessionRelease(ctx.Response())

	if u, ok := sess.Get(global.SESSION_USER_INFO).(model.User); !ok {
		return ctx.JSON(http.StatusOK, map[string]string{
			"code": convert.ToString(http.StatusMovedPermanently),
			"msg":  global.CONTROLLER_USER_LOGIN_FIRST,
		})
	} else {
		return ctx.JSON(http.StatusOK, map[string]string{
			"code":     global.CONTROLLER_CODE_SUCCESS,
			"username": u.Name,
			"phone":    u.Phone,
		})
	}
}

// 退出登录
func (this *User) LoginOut(ctx *echo.Context) error {
	sess, err := session.GetSession(ctx)
	if err != nil {
		log.Error("获取session失败：", err)
		return err
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
	res, _ := this.Base.IsLogin(ctx)
	if !res {
		return nil
	}

	var um = model.User{}
	id := this.Base.GetUserInfo(ctx).Id
	um = um.UserInfoById(id)
	log.Error(um.Avatar)

	return ctx.Render(200, "usercenter", um)
}

// 获取手机验证码
func (this *User) GetUserPhoneCode(ctx *echo.Context) error {
	res, _ := this.Base.IsLogin(ctx)
	if !res {
		return nil
	}
	sess, err := session.GetSession(ctx)
	if err != nil {
		log.Error("获取session失败：", err)
		return err
	}
	defer sess.SessionRelease(ctx.Response())

	var rdb = redis.Get()
	var rkey = "EDITUSERPHONE_" + this.getIp(ctx)
	var rnum = 10
	var phone = ctx.Form("phone")
	var code = convert.ToString(function.GetRand(1000, 9999))
	var userModel = model.User{}

	defer rdb.Close()

	if phone == "" {
		return ctx.JSON(200, map[string]string{
			"code": "300",
			"msg":  "手机号码为空",
		})
	}

	if userModel.PhoneExist(phone) {
		return ctx.JSON(200, map[string]string{
			"code": "300",
			"msg":  "手机号码已被注册",
		})
	}

	num, err := oredis.Int(rdb.Do("GET", rkey))
	if err != nil && err != oredis.ErrNil {
		return ctx.JSON(200, map[string]string{
			"code": "300",
			"msg":  "系统发生错误,稍后重试",
		})
	}
	if num >= rnum {
		return ctx.JSON(200, map[string]string{
			"code": "300",
			"msg":  "今日修改次数过多,请明天再试",
		})
	}

	Sms.SendMsg(phone, "【千兆浏览器】"+code+
		"（验证码）（千兆浏览器客服绝不会索取此验证码，请勿将此验证码告知他人）")

	rdb.Do("INCR", rkey)
	if num == 0 {
		rdb.Do("EXPIRE", rkey, 86400)
	}

	sess.Set("USER_CODE_"+phone, code)

	return ctx.JSON(200, map[string]string{
		"code": "200",
		"msg":  "",
	})
}

// 获取手机验证码
func (this *User) GetPhoneCode(ctx *echo.Context) error {
	sess, err := session.GetSession(ctx)
	if err != nil {
		log.Error("获取session失败：", err)
		return err
	}
	defer sess.SessionRelease(ctx.Response())

	var rdb = redis.Get()
	var rkey = "GETCODE_" + this.getIp(ctx)
	var rnum = 10
	var phone = ctx.Query("phone")
	var code = convert.ToString(function.GetRand(1000, 9999))
	var userModel = model.User{}

	defer rdb.Close()

	if phone == "" {
		return ctx.JSON(200, map[string]string{
			"code": "300",
			"msg":  "手机号码为空",
		})
	}

	if userModel.PhoneExist(phone) {
		return ctx.JSON(200, map[string]string{
			"code": "300",
			"msg":  "手机号码已被注册",
		})
	}

	num, err := oredis.Int(rdb.Do("GET", rkey))

	if err != nil && err != oredis.ErrNil {
		return ctx.JSON(200, map[string]string{
			"code": "300",
			"msg":  "系统发生错误,稍后重试" + err.Error(),
		})
	}
	if num >= rnum {
		return ctx.JSON(200, map[string]string{
			"code": "300",
			"msg":  "今日获取验证码次数过多,请明天再试",
		})
	}

	Sms.SendMsg(phone, "【千兆浏览器】"+code+
		"（验证码）（千兆浏览器客服绝不会索取此验证码，请勿将此验证码告知他人）")

	sess.Set("REG_CODE_"+phone, code)

	rdb.Do("INCR", rkey)
	if num == 0 {
		rdb.Do("EXPIRE", rkey, 86400)
	}
	return ctx.JSON(200, map[string]string{
		"code": "200",
		"msg":  code,
	})
}

// 速度测试
func (this *User) SpeedTest(ctx *echo.Context) error {
	return ctx.Render(200, "speedtest", "")
}

// 验证宽带
func (this *User) VerifyBandwith(ctx *echo.Context) error {
	var (
		action       = ctx.Form("action")
		bandwith     = ctx.Form("bandwith")
		bandwith_pwd = ctx.Form("bandwith_pwd")
		umodel       = model.User{}
	)

	if !grab.In_Array([]string{"verify", "fetch"}, action) {
		return ctx.JSON(http.StatusOK, map[string]string{
			"code": global.CONTROLLER_CODE_ERROR,
			"data": "NULL",
			"msg":  global.CONTROLLER_USER_BANDWITH_ACTIONERROR,
		})
	}

	if bandwith == "" {
		return ctx.JSON(http.StatusOK, map[string]string{
			"code": global.CONTROLLER_CODE_ERROR,
			"data": "NULL",
			"msg":  global.CONTROLLER_USER_BANDWITH_BANDWITHERROR,
		})
	}

	if bandwith_pwd == "" {
		return ctx.JSON(http.StatusOK, map[string]string{
			"code": global.CONTROLLER_CODE_ERROR,
			"data": "NULL",
			"msg":  global.CONTROLLER_USER_BANDWITH_BANDWITHPWDERROR,
		})
	}

	switch action {
	case "verify":
		app_uid := umodel.VerifyBandWith(bandwith, bandwith_pwd)
		return ctx.JSON(http.StatusOK, map[string]interface{}{
			"code": global.CONTROLLER_CODE_SUCCESS,
			"data": map[string]string{"app_uid": app_uid},
			"msg":  global.CONTROLLER_USER_BANDWITH_NOMESSAGE,
		})
	}

	return nil
}

// 获取宽带
func (this *User) GetBandwith(ctx *echo.Context) error {
	var (
		app_uid = ctx.Query("app_uid")
		umodel  = model.User{}
	)

	if app_uid == "" {
		return ctx.JSON(http.StatusOK, map[string]string{
			"code": global.CONTROLLER_CODE_ERROR,
			"msg":  global.CONTROLLER_USER_BANDWITH_NEEDAPPUID,
		})
	}

	u := umodel.GetBrandWith(app_uid)
	if u.Id != "" {
		return ctx.JSON(http.StatusOK, map[string]string{
			"code":     global.CONTROLLER_CODE_SUCCESS,
			"bandwith": u.Bandwith,
		})
	} else {
		return ctx.JSON(http.StatusOK, map[string]string{
			"code": global.CONTROLLER_CODE_ERROR,
			"msg":  global.CONTROLLER_USER_BANDWITH_NOTEXISTBRAND,
		})
	}
}

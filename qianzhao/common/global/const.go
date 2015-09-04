// 定义常量
package global

// session
const (
	SESSION_USER_INFO = "userinfo"
)

// 错误码
const (
	CONTROLLER_CODE_ERROR   = "3xx"
	CONTROLLER_CODE_SUCCESS = "200"
)

// USER
const (
	CONTROLLER_USER_NOT_EXIST            = "用户名不存在,请注册或者联系客服!"
	CONTROLLER_USER_LOGIN_SUCCESS        = "用户登录成功"
	CONTROLLER_USER_CHECK_ERROR          = "验证出错!"
	CONTROLLER_USER_USERPWD_ERROR        = "用户名或密码错误!"
	CONTROLLER_USER_USERNAME_EXIST_ERROR = "该邮箱已被注册使用!"
	CONTROLLER_USER_REGISTER_REF_SUCCESS = "注册或关联成功!"
	CONTROLLER_USER_REGISTER_REF_ERROR   = "一些不知名的错误,注册或关联失败!"
	CONTROLLER_USER_REGISTER_SUCCESS     = "注册成功!"
	CONTROLLER_USER_REGISTER_ERROR       = "注册失败!"
	CONTROLLER_USER_LOGIN_FIRST          = "请先登陆!"
	CONTROLLER_USER_LOGINOUT             = "用户已注销!"
	CONTROLLER_USER_NEEDLOGIN            = "need login!"
)

// Favorite
const (
	CONTROLLER_FAVORITE_NOCONTENT       = "无内容!"
	CONTROLLER_FAVORITE_NOUPLOADCONTENT = "没有需要上传的内容!"
)

// ebit
const (
	CONTROLLER_EBIT_NOPARAM            = "缺少参数!"
	CONTROLLER_EBIT_REQUESTFAILE       = "请求失败!"
	CONTROLLER_EBIT_SPEEDTFAILE        = "提速失败!!!"
	CONTROLLER_EBIT_REQUEST_TIMEOUT    = "抱歉，本次请求超时，请稍后重试"
	CONTROLLER_EBIT_NOSPEEDUPCONDITION = "抱歉，您的运行环境不符合加速条件"
	CONTROLLER_EBIT_SIGNERROR          = "sign错误!"
)

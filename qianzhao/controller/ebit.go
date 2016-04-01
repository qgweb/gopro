package controller

import (
	"encoding/json"
	"github.com/ngaut/log"

	"net/http"
	"time"

	"github.com/qgweb/gopro/lib/convert"

	"github.com/astaxie/beego/httplib"
	oredis "github.com/garyburd/redigo/redis"
	"github.com/labstack/echo"
	"github.com/qgweb/gopro/lib/encrypt"
	"github.com/qgweb/gopro/qianzhao/common/function"
	"github.com/qgweb/gopro/qianzhao/common/global"
	"github.com/qgweb/gopro/qianzhao/common/redis"
)

const (
	APP_KEY       = "APP_MSB704PU"
	APP_SECRET    = "mv6oy8f2qo0l0ogvxnm02tM7"
	EBIT_BASE_URL = "http://218.85.118.9:8000/api2/"
)

type Ebit struct {
}

// 应用程序鉴权
func getSecret(timestamp string) string {
	return encrypt.DefaultMd5.Encode(APP_KEY + timestamp + APP_SECRET)
}

// 提速
func (this *Ebit) SpeedupOpen(ctx *echo.Context) error {
	var (
		sid  = function.GetPost(ctx, "sid")
		conn = redis.Get()
	)
	defer conn.Close()

	speedup_data, err := oredis.String(conn.Do("GET", sid))

	if err != oredis.ErrNil {
		log.Warn("[ebit speedupopen] redis获取可以失败 ", err)
		return ctx.String(http.StatusInternalServerError, "系统发生错误")
	}

	speedup_dataJson := make(map[string]interface{})
	if err := json.Unmarshal([]byte(speedup_data), &speedup_dataJson); err != nil {
		log.Warn("[ebit speedupopen] 解析json出错", err)
		return ctx.String(http.StatusInternalServerError, "系统发生错误")
	}

	dial_acct := ""
	ip := ""

	if v, ok := speedup_dataJson["dial_acct"]; ok {
		dial_acct = convert.ToString(v)
	}

	if v, ok := speedup_dataJson["ip"]; ok {
		ip = convert.ToString(v)
	}

	timestamp := function.GetTimeUnix()
	secret := getSecret(timestamp)

	req := httplib.Post(EBIT_BASE_URL + "speedup/opn")
	req.JSONBody(map[string]string{
		"app":       APP_KEY,
		"secret":    secret,
		"timestamp": timestamp,
		"ip_addr":   ip,
		"duration":  "60",
		"dial_acct": dial_acct,
	})
	res := make(map[string]interface{})
	req.ToJSON(&res)

	time.Sleep(time.Second * 3)
	if v, ok := res["task_id"]; ok {
		queryRes := TaskQuery(convert.ToString(v))
		message := ""
		if v, ok := queryRes["message"]; ok {
			message = convert.ToString(v)
		}

		if v, ok := queryRes["errno"]; !ok || convert.ToInt(v) != 0 {
			return ctx.JSON(http.StatusOK, map[string]string{
				"code": global.CONTROLLER_CODE_ERROR,
				"msg":  message,
			})
		} else {
			start_time := function.GetTimeUnix()
			speedup_dataJson["start_time"] = start_time
			key := "HH_" + ip

			d, err := json.Marshal(&speedup_dataJson)
			if err != nil {
				log.Warn("[ebit speedupopen] 解析json出错", err)
				return ctx.String(http.StatusInternalServerError, "系统发生错误")
			}

			conn.Do("SET", key, string(d))

			return ctx.JSON(http.StatusOK, map[string]string{
				"code":       global.CONTROLLER_CODE_SUCCESS,
				"msg":        message,
				"start_time": start_time,
			})
		}
	}

	return nil
}

// 用户查询 判断是否满足提速条件
func (this *Ebit) SpeedupOpenCheck(ctx *echo.Context) error {
	var (
		ip   = function.GetIP(ctx)
		conn = redis.Get()
		key  = "HH_" + ip
	)

	defer conn.Close()

	speedup_data, err := oredis.String(conn.Do("GET", key))
	if err != oredis.ErrNil {
		log.Warn("[ebit speedupOpencheck] 读取redis失败", err)
	}

	if speedup_data != "" {
		speedup_dataJson := make(map[string]interface{})
		err := json.Unmarshal([]byte(speedup_data), &speedup_dataJson)
		if err != nil {
			log.Warn("[ebit speedupOpencheck] 解析speedup_data数据出错", err)
			return ctx.String(http.StatusInternalServerError, "系统发生错误")
		}

		if start_time, ok := speedup_dataJson["start_time"]; ok {
			return ctx.JSON(http.StatusOK, map[string]string{
				"code":       global.CONTROLLER_CODE_SUCCESS,
				"start_time": convert.ToString(start_time),
				"cur_time":   function.GetTimeUnix(),
			})
		} else {
			return ctx.JSON(http.StatusOK, map[string]string{
				"code": global.CONTROLLER_CODE_SUCCESS,
				"sid":  key,
			})
		}
	}

	timestamp := function.GetTimeUnix()
	secret := getSecret(timestamp)

	req := httplib.Post(EBIT_BASE_URL + "user/query")
	req.JSONBody(map[string]string{
		"app":       APP_KEY,
		"secret":    secret,
		"timestamp": timestamp,
		"_type":     "0",
		"data":      ip,
	})

	res := make(map[string]interface{})
	req.ToJSON(&res)

	if _, ok := res["task_id"]; !ok {
		return ctx.JSON(http.StatusOK, map[string]string{
			"code":    global.CONTROLLER_CODE_ERROR,
			"message": global.CONTROLLER_EBIT_REQUEST_TIMEOUT,
		})
	}

	queryRes := TaskQuery(convert.ToString(res["task_id"]))

	if v, ok := queryRes["errno"]; !ok || convert.ToInt(v) != 0 {
		return ctx.JSON(http.StatusOK, map[string]string{
			"code":    global.CONTROLLER_CODE_ERROR,
			"errno":   convert.ToString(v),
			"message": global.CONTROLLER_EBIT_NOSPEEDUPCONDITION,
		})
	} else {
		value := map[string]string{
			"ip":            ip,
			"dial_acct":     convert.ToString(queryRes["dial_acct"]),
			"speedup_check": "1",
		}

		rv, err := json.Marshal(value)
		if err != nil {
			return ctx.String(http.StatusInternalServerError, "系统发生错误")
		}

		conn.Do("SET", key, string(rv))

		return ctx.JSON(http.StatusOK, map[string]string{
			"code": global.CONTROLLER_CODE_SUCCESS,
			"sid":  key,
		})
	}

	return nil
}

// 用户当前提速状态查询
func (this *Ebit) SpeedupCheck(ctx *echo.Context) error {
	var (
		ip   = function.GetPost(ctx, "ip")
		sign = function.GetPost(ctx, "sign")
		key  = ip + "_h" + sign
		conn = redis.Get()
	)

	defer conn.Close()

	if ip == "" || sign == "" {
		return ctx.JSON(http.StatusOK, map[string]string{
			"code": global.CONTROLLER_CODE_ERROR,
			"msg":  global.CONTROLLER_EBIT_NOPARAM,
		})
	}

	dial_acct, err := oredis.String(conn.Do("GET", key))
	if err != oredis.ErrNil {
		log.Warn("[ebit speedupopen] redis获取可以失败 ", err)
		return ctx.String(http.StatusInternalServerError, "系统发生错误")
	}

	if dial_acct == "" {
		return ctx.JSON(http.StatusOK, map[string]string{
			"code": global.CONTROLLER_CODE_ERROR,
			"msg":  global.CONTROLLER_EBIT_SIGNERROR,
		})
	}

	timestamp := function.GetTimeUnix()
	secret := getSecret(timestamp)

	req := httplib.Post(EBIT_BASE_URL + "speedup/check")
	req.JSONBody(map[string]string{
		"app":       APP_KEY,
		"secret":    secret,
		"timestamp": timestamp,
		"ip_addr":   ip,
		"dial_acct": dial_acct,
	})
	res := make(map[string]interface{})
	req.ToJSON(&res)

	if v, ok := res["task_id"]; !ok {
		return ctx.JSON(http.StatusOK, map[string]string{
			"code": global.CONTROLLER_CODE_ERROR,
			"msg":  global.CONTROLLER_EBIT_REQUESTFAILE,
		})
	} else {
		time.Sleep(time.Second * 5)
		queryRes := TaskQuery(convert.ToString(v))

		if v, ok := queryRes["errno"]; !ok || convert.ToInt(v) != 0 {
			return ctx.JSON(http.StatusOK, map[string]string{
				"code": global.CONTROLLER_CODE_ERROR,
			})
		} else {
			return ctx.JSON(http.StatusOK, map[string]string{
				"code":      global.CONTROLLER_CODE_SUCCESS,
				"dial_acct": dial_acct[0:4],
			})
		}
	}
}

func TaskQuery(taskId string) map[string]interface{} {
	var (
		timestamp = function.GetTimeUnix()
		secret    = getSecret(timestamp)
	)

	req := httplib.Post(EBIT_BASE_URL + "task/query")

	body := make(map[string]string)
	body["app"] = APP_KEY
	body["secret"] = secret
	body["timestamp"] = timestamp
	body["task_id"] = taskId

	req.JSONBody(&body)

	v := make(map[string]interface{})

	req.ToJSON(&v)
	return v
}

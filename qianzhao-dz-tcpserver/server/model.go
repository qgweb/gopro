// 具体业务罗辑
package server

import (
	"time"

	"github.com/qgweb/gopro/lib/convert"

	"github.com/astaxie/beego/httplib"
	"github.com/ngaut/log"
	"github.com/qgweb/gopro/lib/encrypt"
	"github.com/qgweb/gopro/qianzhao-dz-tcpserver/model"
)

const (
	APP_KEY       = "APP_MSB704PU"
	APP_SECRET    = "mv6oy8f2qo0l0ogvxnm02tM7"
	EBIT_BASE_URL = "http://218.85.118.9:8000/api2/"
)

type BDInterfaceManager struct{}

func (this *BDInterfaceManager) HaveTime(account string) int {
	model := model.BrandAccountRecord{}
	return model.GetAccountCanUserTime(account)
}

// 判断是否开启
func (this *BDInterfaceManager) CanStart(ip string) string {
	var (
		timestamp = convert.ToString(time.Now().Unix())
		secret    = getSecret(timestamp)
	)

	// 正式删除
	//return "1111"

	req := httplib.Post(EBIT_BASE_URL + "user/query")
	req.JsonBody(map[string]string{
		"app":       APP_KEY,
		"secret":    secret,
		"timestamp": timestamp,
		"_type":     "0",
		"data":      ip,
	})

	res := make(map[string]interface{})
	req.ToJson(&res)
	if CheckError(res) {
		return ""
	}

	if task_id, ok := res["task_id"]; ok {
		time.Sleep(time.Second * 3)
		info := TaskQuery(task_id.(string))
		if CheckError(info) {
			return ""
		}
		return info["dial_acct"].(string)
	}

	return ""

}

// 开启
func (this *BDInterfaceManager) Start(account string, ip string) Respond {
	var (
		timestamp = convert.ToString(time.Now().Unix())
		secret    = getSecret(timestamp)
		errmsg    = "抱歉，程序发生错误,提速失败"
	)

	//// 正式删除
	//return Respond{Code: "200", Msg: "xxxxxxxxxxx"}

	req := httplib.Post(EBIT_BASE_URL + "speedup/open")
	req.JsonBody(map[string]string{
		"app":       APP_KEY,
		"secret":    secret,
		"timestamp": timestamp,
		"ip_addr":   ip,
		"duration":  "60",
		"dial_acct": account,
	})
	res := make(map[string]interface{})
	req.ToJson(&res)

	if CheckError(res) {
		return Respond{Code: "500", Msg: errmsg}
	}

	if task_id, ok := res["task_id"]; ok {
		time.Sleep(time.Second * 3)
		info := TaskQuery(task_id.(string))
		if CheckError(info) {
			return Respond{Code: "500", Msg: errmsg}
		}
		return Respond{Code: "200", Msg: info["channel_id"].(string)}
	}

	return Respond{Code: "500", Msg: errmsg}
}

// 关闭
func (this *BDInterfaceManager) Stop(channel_id string) Respond {
	var (
		timestamp = convert.ToString(time.Now().Unix())
		secret    = getSecret(timestamp)
		errmsg    = "抱歉，程序发生错误,提速关闭失败"
	)
	//// 正式删除
	//return Respond{Code: "200", Msg: ""}

	req := httplib.Post(EBIT_BASE_URL + "speedup/open")
	req.JsonBody(map[string]string{
		"app":        APP_KEY,
		"secret":     secret,
		"timestamp":  timestamp,
		"channel_id": channel_id,
	})
	res := make(map[string]interface{})
	req.ToJson(&res)

	if CheckError(res) {
		return Respond{Code: "500", Msg: errmsg}
	}

	if task_id, ok := res["task_id"]; ok {
		time.Sleep(time.Second * 3)
		info := TaskQuery(task_id.(string))
		if CheckError(info) {
			return Respond{Code: "500", Msg: errmsg}
		}
		return Respond{Code: "200", Msg: ""}
	}
	return Respond{Code: "500", Msg: errmsg}
}

func CheckError(res map[string]interface{}) bool {
	if err, ok := res["errno"]; ok && err.(float64) != 0 {
		if msg, ok := res["message"]; ok {
			log.Error(msg)
		}
		return true
	}
	return false
}

func TaskQuery(taskId string) map[string]interface{} {
	var (
		timestamp = convert.ToString(time.Now().Unix())
		secret    = getSecret(timestamp)
	)

	req := httplib.Post(EBIT_BASE_URL + "task/query")

	body := make(map[string]string)
	body["app"] = APP_KEY
	body["secret"] = secret
	body["timestamp"] = timestamp
	body["task_id"] = taskId

	req.JsonBody(&body)

	v := make(map[string]interface{})

	req.ToJson(&v)
	return v
}

func getSecret(timestamp string) string {
	return encrypt.DefaultMd5.Encode(APP_KEY + timestamp + APP_SECRET)
}

// http 各种接口
// sms短信接口
package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/astaxie/beego/httplib"
	"github.com/bradfitz/gomemcache/memcache"
	"github.com/qgweb/gopro/lib/convert"
	"gopkg.in/ini.v1"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
)

var (
	conf    = flag.String("conf", "conf.ini", "配置文件")
	iniFile *ini.File
	err     error
)

func init() {
	data, err := ioutil.ReadFile(*conf)
	if err != nil {
		log.Fatalln("打开配置文件失败,错误信息:", err)
	}

	iniFile, err = ini.Load(data)
	if err != nil {
		log.Fatalln("加载配置文件内容失败,错误信息:", err)
	}
}

// 发送短信接口
func sms(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	var (
		// 短信接口
		smsuser     = iniFile.Section("sms").Key("user").String()
		smspwd      = iniFile.Section("sms").Key("pwd").String()
		smsurl      = iniFile.Section("sms").Key("url").String()
		bckuser     = iniFile.Section("sms").Key("ckuser").String()
		smstimes, _ = iniFile.Section("sms").Key("times").Int()

		// 短信接口参数
		mobile   = r.PostForm.Get("mobile")
		content  = r.PostForm.Get("content")
		extcode  = r.PostForm.Get("extcode")
		senddate = r.PostForm.Get("senddate")
		batchID  = r.PostForm.Get("batchID")
		platform = r.PostForm.Get("platform")
		ckuser   = r.PostForm.Get("ckuser")
	)

	if ckuser != bckuser {
		w.Write([]byte("拒绝访问"))
		return
	}

	// 验证手机号码不能发送多次
	err = checkMulMobileSend(mobile, platform, smstimes, func() error {
		//$this->url = 'http://116.213.72.20/SMSHttpService/send.aspx';
		//$post_data = array();
		// $post_data['username'] = $this->username;//用户名
		// $post_data['password'] = $this->password;//密码
		// $post_data['mobile'] = $mobile;//手机号，多个号码以分号分隔，如：13407100000;13407100001;13407100002
		// $post_data['content'] = urlencode("【聚惠算】{$info}，验证码：{$content}，请勿向他人泄露此验证码");//内容，如为中文一定要使用一下urlencode函数
		// $post_data['extcode'] = "";//扩展号，可选
		// $post_data['senddate'] = "";//发送时间，格式：yyyy-MM-dd HH:mm:ss，可选
		// $post_data['batchID'] = "";//批次号，可选
		req := httplib.Post(smsurl)
		req.Param("username", smsuser)
		req.Param("password", smspwd)
		req.Param("mobile", mobile)
		req.Param("content", content)
		req.Param("extcode", extcode)
		req.Param("senddate", senddate)
		req.Param("batchID", batchID)

		res, err := req.String()
		if err != nil {
			return err
		}

		if res != "0" {
			log.Println(res)
			return errors.New("发送短信失败")
		}

		return nil
	})

	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}

	w.Write([]byte("0"))
}

// 验证手机号码多次发送
func checkMulMobileSend(phone string, platform string, times int, sendfun func() error) error {
	var (
		memhost = iniFile.Section("memcache").Key("host").String()
		memport = iniFile.Section("memcache").Key("port").String()
		key     = platform + "_REG_PHONE_TIMES_" + phone
	)

	//连接memcache
	client := memcache.New(fmt.Sprintf("%s:%s", memhost, memport))
	it, err := client.Get(key)
	count := 0

	if err != nil && err != memcache.ErrCacheMiss {
		log.Println("memcach获取数据失败:", err)
		return errors.New("连接memcache失败")
	}

	if it != nil {
		count = convert.ToInt(string(it.Value))
	}

	if count >= times {
		return errors.New("手机注册发送验证码一天之内不能超过" + strconv.Itoa(times) + "次")
	}

	//发送短信
	if err = sendfun(); err != nil {
		return errors.New("发送短信失败")
	}

	//存入memcache
	count = count + 1
	item := &memcache.Item{}
	item.Expiration = 3600 * 24
	item.Key = key
	item.Value = []byte(convert.ToString(count))
	err = client.Set(item)
	if err != nil {
		log.Println("memcache设置失败,错误信息为:", err)
		return errors.New("设置memcache key失败")
	}
	return nil
}

func main() {
	var (
		host = iniFile.Section("http").Key("host").String()
		port = iniFile.Section("http").Key("port").String()
	)

	//发送短信接口
	http.HandleFunc("/SMSHttpService/send.aspx", sms)
	log.Fatalln(http.ListenAndServe(fmt.Sprintf("%s:%s", host, port), nil))
}

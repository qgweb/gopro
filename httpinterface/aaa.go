package main

import (
	"fmt"
	"net/url"

	"github.com/astaxie/beego/httplib"
	//"github.com/bradfitz/gomemcache/memcache"
	//"time"
)

func main() {
	//$this->url = 'http://116.213.72.20/SMSHttpService/send.aspx';
	//$post_data = array();
	// $post_data['username'] = $this->username;//用户名
	// $post_data['password'] = $this->password;//密码
	// $post_data['mobile'] = $mobile;//手机号，多个号码以分号分隔，如：13407100000;13407100001;13407100002
	// $post_data['content'] = urlencode("【聚惠算】{$info}，验证码：{$content}，请勿向他人泄露此验证码");//内容，如为中文一定要使用一下urlencode函数
	// $post_data['extcode'] = "";//扩展号，可选
	// $post_data['senddate'] = "";//发送时间，格式：yyyy-MM-dd HH:mm:ss，可选
	// $post_data['batchID'] = "";//批次号，可选
	req := httplib.Post("http://122.225.98.72:9001/SMSHttpService/send.aspx")
	req.Param("username", "jhsyzm")
	req.Param("password", "jhsyzmx")
	req.Param("mobile", "15158117079")
	req.Param("content", url.QueryEscape("【聚惠算】test4，验证码：11036，请勿向他人泄露此验证码"))
	req.Param("extcode", "")
	req.Param("senddate", "")
	req.Param("batchID", "")
	req.Param("platform", "JHS")
	req.Param("ckuser", "qgadmin")
	fmt.Println(req.String())
}

package main

import (
	"fmt"
	"github.com/bitly/go-simplejson"
	"github.com/juju/errors"
	"github.com/qgweb/gopro/xurpc/common"
	"github.com/qgweb/new/lib/convert"
	"github.com/qgweb/new/lib/timestamp"
	"net/url"
	"github.com/astaxie/beego/httplib"
)

// 统计接口数据
type Stats struct {
}

// 获取标签的数量
// 可根据天来获取
func (this Stats) GetTagCount(province string, tp string, tagid string, day int) []byte {
	config, err := common.GetBeegoIniObject()
	if err != nil {
		return jsonReturn("0", errors.New("配置文件出错"))
	}
	btime := timestamp.GetDayTimestamp(-1 * day)
	etime := timestamp.GetDayTimestamp(0)

	var countFun = func(key string) int64 {
		aurl := config.String("stats::host") + "/api/get"
		u := url.Values{}
		u.Add("db", "tags_stats")
		u.Add("key", key)
		aurl += "?" + u.Encode()
		req := httplib.Get(aurl)
		bs, err := req.Bytes()
		if err != nil {
			return 0
		}
		if sj, err := simplejson.NewJson(bs); err == nil {
			var count = int64(0)
			if ret, err := sj.Get("ret").String(); ret == "0" && err == nil {
				if d, err := sj.Get("data").String(); err == nil {
					count += convert.ToInt64(d)
				}
				return count
			}
		}
		return 0
	}

	count := int64(0)
	for t := convert.ToInt64(btime); t < convert.ToInt64(etime); t += 3600 {
		kk := fmt.Sprintf("%s_%s_%s_%s", province, convert.ToString(t), tp, tagid)
		count += countFun(kk)
	}
	return jsonReturn(count, nil)
}
//http://qy.weijiabao.com.cn/index.php/qy/ThZhVf1463104885
//ThZhVf1463104885
//XZU2Yrqon6zu2XW5gsCnVAd5mOPqp528qTVqec7ZcuG

//http://cqy.weijiabao.com.cn/index.php/qy/ZhuLJM1470301430
//ZhuLJM1470301430
//O7nT5decFcPuy6gQaxywCaPfReWLq6YEDLX8pBKue0p

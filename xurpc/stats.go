package main

import (
	"fmt"
	"github.com/bitly/go-simplejson"
	"github.com/juju/errors"
	"github.com/qgweb/gopro/xurpc/common"
	"github.com/qgweb/new/lib/convert"
	"github.com/qgweb/new/lib/timestamp"
	"io/ioutil"
	"net/http"
	"net/url"
)

// 统计接口数据
type Stats struct {
}

// 获取标签的数量
// 可根据天来获取
func (this Stats) GetTagCount(tp string, tagid string, day int) []byte {
	config, err := common.GetBeegoIniObject()
	if err != nil {
		return jsonReturn("0", errors.New("配置文件出错"))
	}
	btime := timestamp.GetDayTimestamp(-1 * day)
	etime := timestamp.GetHourTimestamp(-1)

	aurl := config.String("stats::host") + "/api/list"
	u := url.Values{}
	u.Add("db", "tags_stats")
	u.Add("bkey", fmt.Sprintf("zj_%s_%s_%s", btime, tp, tagid))
	u.Add("ekey", fmt.Sprintf("zj_%s_%s_%s", etime, tp, tagid))
	u.Add("limit", "10000000")
	aurl += "?" + u.Encode()
	resp, err := http.Get(aurl)
	if err != nil {
		return jsonReturn("0", errors.New("接口获取失败"))
	}
	if resp != nil && resp.Body != nil {
		defer resp.Body.Close()
		info, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return jsonReturn("0", errors.New("接口获取失败"))
		}
		if sj, err := simplejson.NewJson(info); err == nil {
			var count = 0
			if ret, err := sj.Get("ret").String(); ret == "0" && err == nil {
				if data, err := sj.Get("data").Map(); err == nil {
					for _, v := range data {
						if vv, ok := v.(map[string]interface{}); ok {
							count += convert.ToInt(vv["Value"])
						}
					}
					return jsonReturn(count, nil)
				}
			}
		}
	}
	return jsonReturn("0", nil)
}

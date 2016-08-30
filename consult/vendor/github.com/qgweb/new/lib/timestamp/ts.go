package timestamp

import (
	"time"

	"github.com/qgweb/new/lib/convert"
)

// 获取当前时间戳
// 2006-01-02 15:04:05
func GetTimestamp(v ...interface{}) string {
	if len(v) > 0 {
		a,_:=time.ParseInLocation("2006-01-02 15:04:05",v[0].(string),time.Local)
		return convert.ToString(a.Unix())
	}
	return convert.ToString(time.Now().Unix())
}


// 获取某个小时的时间戳
func GetHourTimestamp(hour int) string {
	d := time.Now().Add(time.Hour * time.Duration(hour)).Format("2006010215")
	a, _ := time.ParseInLocation("2006010215", d, time.Local)
	return convert.ToString(a.Unix())
}

// 获取任意一天整点时间戳
func GetDayTimestamp(day int) string {
	d := time.Now().AddDate(0, 0, day).Format("20060102")
	a, _ := time.ParseInLocation("20060102", d, time.Local)
	return convert.ToString(a.Unix())
}

// 获取月的时间戳
func GetMonthTimestamp(month int) string {
	d := time.Now().AddDate(0, month, 0).Format("200601")
	a, _ := time.ParseInLocation("200601", d, time.Local)
	return convert.ToString(a.Unix())
}

// unix转日期
func GetUnixFormat(unix string) string{
	return time.Unix(convert.ToInt64(unix),0).Format("2006-01-02 15:04:05")
}
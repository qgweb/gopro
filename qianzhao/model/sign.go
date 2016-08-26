package model

import (
	"github.com/qgweb/new/lib/convert"
	"github.com/qgweb/new/lib/timestamp"
	"time"
	"strings"
)

//签到表
type Sign struct {
	Id      int
	Uid     int
	Btime   int64
	History string
}

const (
	SIGN_TABLE_NAME = "221su_sign"
)

func (this *Sign) To(v map[string]string) (ar Sign) {
	if vv, ok := v["id"]; ok {
		ar.Id = convert.ToInt(vv)
	}
	if vv, ok := v["uid"]; ok {
		ar.Uid = convert.ToInt(vv)
	}
	if vv, ok := v["history"]; ok {
		ar.History = vv
	}
	if vv, ok := v["btime"]; ok {
		ar.Btime = convert.ToInt64(vv)
	}
	return
}

func (this *Sign) GetInfo(uid int) (s Sign, err error) {
	sql := myorm.BSQL().Select("*").From(SIGN_TABLE_NAME).Where("uid=?").GetSQL()
	list, err := myorm.Get(sql, uid)
	if err != nil {
		return s, err
	}
	return this.To(list), nil
}

func (this *Sign) setHistory(btime int64, histroy string) (string, int64) {
	day := int(time.Unix(convert.ToInt64(timestamp.GetDayTimestamp(0)), 0).Sub(time.Unix(btime, 0)).Hours() / 24)
	hs := strings.Split(histroy, "")
	hs[day] = "1"
	if day > 0 && hs[day - 1] == "0" {
		return "10000", convert.ToInt64(timestamp.GetDayTimestamp(0))
	}

	return strings.Join(hs, ""), btime
}

func (this *Sign) getHistory(btime int64, histroy string) bool {
	day := int(time.Unix(convert.ToInt64(timestamp.GetDayTimestamp(0)), 0).Sub(time.Unix(btime, 0)).Hours() / 24)
	hs := strings.Split(histroy, "")
	if day >= 5 {
		return false
	}
	return hs[day] == "1"
}

func (this *Sign) HasSign(uid int) (bool) {
	info, err := this.GetInfo(uid)
	if err != nil {
		return false
	}
	return info.History != "" && this.getHistory(info.Btime, info.History)
}

func (this *Sign) Add(uid int, fun func()) (string, error) {
	info, err := this.GetInfo(uid)
	if err != nil {
		return "00000", err
	}

	if info.Id == 0 {
		//添加
		sql := myorm.BSQL().Insert(SIGN_TABLE_NAME).Values("uid", "btime", "history").GetSQL()
		_, err := myorm.Insert(sql, uid, timestamp.GetDayTimestamp(0), "1" + strings.Repeat("0", 4))
		return "10000", err
	} else {
		//修改
		if ok, _ := this.Reset(uid); ok {
			info.History = "00000"
			info.Btime = convert.ToInt64(timestamp.GetDayTimestamp(0))
		}
		sql := myorm.BSQL().Update(SIGN_TABLE_NAME).Set("history", "btime").Where("uid=?").GetSQL()
		his, bt := this.setHistory(info.Btime, info.History)
		_, err := myorm.Update(sql, his, bt, uid)
		if strings.Count(his, "1") == 5 {
			fun()
		}
		return his, err
	}
}

func (this *Sign) Reset(uid int) (bool, error) {
	info, err := this.GetInfo(uid)
	if err != nil || info.Id == 0 {
		return false, err
	}
	var bt int64 = info.Btime
	var his string = info.History
	day := int(time.Unix(convert.ToInt64(timestamp.GetDayTimestamp(0)), 0).Sub(time.Unix(info.Btime, 0)).Hours() / 24)
	hs := strings.Split(info.History, "")

	if (day > 0 && hs[day - 1] == "0") || (strings.Count(info.History, "1") == 5 && day >= 5) {
		his = "00000"
		bt = convert.ToInt64(timestamp.GetDayTimestamp(0))
	}

	sql := myorm.BSQL().Update(SIGN_TABLE_NAME).Set("btime", "history").Where("uid=?").GetSQL()
	n, err := myorm.Update(sql, bt, his, uid)
	return n > 0, err
}




package model

import (
	"github.com/qgweb/new/lib/convert"
	"github.com/qgweb/new/lib/timestamp"
)

type Word struct {
	Id     int
	Pic    string
	Word   string
	Title  string
	Status int
	Time   int
}

const (
	WORD_TABLE_NAME = "221su_word"
)

func (this *Word) Get() (w Word, err error) {
	sql := myorm.BSQL().Select("*").From(WORD_TABLE_NAME).Where("time=?").GetSQL()
	info, err := myorm.Query(sql, convert.ToInt(timestamp.GetHourTimestamp(0)))
	if err != nil {
		return w, err
	}
	if (len(info) > 0) {
		w.Id = convert.ToInt(info[0]["id"])
		w.Pic = info[0]["pic"]
		w.Word = info[0]["word"]
		w.Title = info[0]["title"]
		w.Status = convert.ToInt(info[0]["status"])
		w.Time = convert.ToInt(info[0]["time"])
		return w, nil
	}
	return
}

func (this *Word) Has(uid int) (bool, error) {
	sql := myorm.BSQL().Select("id").From("221su_award_record").
		Where("time=? and user_id=? and source=?").GetSQL()
	info, err := myorm.Get(sql, timestamp.GetHourTimestamp(0), uid, 1)
	if err != nil {
		return false, err
	}
	if len(info) > 0 {
		return true, nil
	}
	return false, nil
}

func (this *Word) HasCount() (int, error) {
	sql := myorm.BSQL().Select("count(*) as num").From("221su_award_record").
		Where("time=?  and source=? and awards_type!=0").GetSQL()
	info, err := myorm.Get(sql, timestamp.GetHourTimestamp(0), 1)
	if err != nil {
		return 0, err
	}
	if len(info) > 0 {
		return convert.ToInt(info["num"]), nil
	}
	return 0, nil
}

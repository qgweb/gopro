package model

import (
	"database/sql"
	"fmt"
	"github.com/juju/errors"
	"github.com/ngaut/log"
	"github.com/qgweb/gopro/lib/convert"
	"github.com/qgweb/gopro/lib/encrypt"
	"github.com/qgweb/new/lib/timestamp"
	"math/rand"
	"time"
)

type Award struct {
	Id         int
	UserId     string
	AwardsType int
	Code       string
	Time       int
	Date       int
	Source     int
}

const (
	AWARD_TABLE_NAME = "221su_award_record"
	AWARD_CODE_TABLE_NAME = "221su_recharge_code"
)

// 是否已经参加过
func (this *Award) HaveJoin(userid string, date string) (bool, error) {
	sql := myorm.BSQL().Select("count(*) as num").From(AWARD_TABLE_NAME).Where("user_id=? and date=? and source=0").GetSQL()
	list, err := myorm.Query(sql, userid, date)
	if err != nil {
		return false, err
	}
	if len(list) > 0 {
		return convert.ToInt(list[0]["num"]) > 0, nil
	}
	return false, nil
}

func (this *Award) getRand(ary map[int]int) int {
	prosum := 0
	for _, v := range ary {
		prosum += v
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	for k, v := range ary {
		randNum := r.Intn(prosum)
		if randNum <= v {
			return k
		} else {
			prosum -= v
		}
	}

	return -1
}

func (this *Award) black(tx *sql.Tx, id string) (bool, error) {
	r, err := tx.Query("select count(*) as num from 221su_users where created>=? and created<? and id=?",
		1458057600, 1460476800, id)
	if err != nil {
		return true, err
	}
	defer r.Close()
	if r.Next() {
		var id int
		err := r.Scan(&id)
		if err != nil {
			return true, err
		}
		if id >= 1 {
			return true, nil
		}
	}
	return false, nil
}

func (this *Award) Get(userid string, awardnum map[int]int, source int) (int, string, error) {
	tx, err := myorm.Begin()
	defer tx.Commit()
	if err != nil {
		return 0, "", err
	}
	// 验证是否黑名单
	b, err := this.black(tx, userid)
	if err != nil || b {
		return 0, "", err
	}

	// 获取抽奖号码
	var n int
	if source == 1 {
		//按小时
		n, err = this.getAwardNumClock(tx, awardnum)
	} else {
		n, err = this.getAwardNum(tx, awardnum)
	}

	if err != nil {
		return 0, "", err
	}

	// 获取充值卡号
	var code = ""
	if n != 0 {
		code, err = this.getAwardCode(tx, n)
	}

	// 添加中奖记录
	info := Award{}
	info.UserId = userid
	info.Time = convert.ToInt(timestamp.GetTimestamp())
	if source == 1 {
		info.Time = convert.ToInt(timestamp.GetHourTimestamp(0))
	}
	info.Date = convert.ToInt(timestamp.GetDayTimestamp(0))
	info.Code = code
	info.AwardsType = n
	info.Source = source
	if this.addReocrd(tx, info) != nil {
		tx.Rollback()
		return 0, "", err
	}
	return n, code, nil
}

func (this *Award) getAwardNum(tx *sql.Tx, awardnum map[int]int) (int, error) {
	row, err := tx.Query("SELECT count(*), awards_type from 221su_award_record where date=? group by awards_type",
		timestamp.GetDayTimestamp(0))
	if err != nil {
		return 0, err
	}

	defer row.Close()

	for row.Next() {
		var count int
		var atype int
		err := row.Scan(&count, &atype)
		if err != nil {
			return 0, err
		}
		log.Info(count, atype)
		if v, ok := awardnum[atype]; ok {
			if v - count <= 0 {
				delete(awardnum, atype)
				awardnum[0] += count
			} else {
				awardnum[atype] -= count
			}
		}
	}

	n := this.getRand(awardnum)
	if n < 0 {
		return 0, nil
	}
	return n, nil
}

func (this *Award) getAwardNumClock(tx *sql.Tx, awardnum map[int]int) (int, error) {
	row, err := tx.Query("SELECT count(*), awards_type from 221su_award_record where time=? group by awards_type",
		timestamp.GetHourTimestamp(0))
	if err != nil {
		return 0, err
	}

	defer row.Close()

	for row.Next() {
		var count int
		var atype int
		err := row.Scan(&count, &atype)
		if err != nil {
			return 0, err
		}
		log.Info(count, atype)
		if v, ok := awardnum[atype]; ok {
			if v - count <= 0 {
				if v != 0 {
					delete(awardnum, atype)
				}
			} else {
				awardnum[atype] -= count
			}
		}
	}
	fmt.Println("award 191", awardnum)
	n := this.getRand(awardnum)
	if n < 0 {
		return 0, nil
	}
	return n, nil
}

func (this *Award) getAwardCode(tx *sql.Tx, t int) (string, error) {
	sql := myorm.BSQL().Select("code", "id").From(AWARD_CODE_TABLE_NAME).Where("status=? and type=?").Limit(1).GetSQL()
	row, err := tx.Query(sql, 0, t)
	if err != nil {
		return "", err
	}
	if row.Next() {
		var code string
		var id int
		err := row.Scan(&code, &id)
		if err != nil {
			return "", err
		}

		//修改状态
		log.Info(code, id)
		row.Close()
		sql = myorm.BSQL().Update(AWARD_CODE_TABLE_NAME).Set("status").Where("id=?").GetSQL()
		r, err := tx.Exec(sql, 1, id)
		log.Info(sql, r, err)
		if err != nil {
			return "", err
		}
		if v, err := r.RowsAffected(); err == nil && v == 1 {
			return code, nil
		}
	}
	return "", nil
}

func (this *Award) addReocrd(tx *sql.Tx, info Award) error {
	sql := myorm.BSQL().Insert(AWARD_TABLE_NAME).Values("user_id", "awards_type", "code", "date", "time", "source").GetSQL()
	r, err := tx.Exec(sql, info.UserId, info.AwardsType, info.Code, info.Date, info.Time, info.Source)
	if err != nil {
		return err
	}
	if _, err := r.LastInsertId(); err == nil {
		return nil
	}
	return errors.New("插入失败")
}

func (this *Award) MyRecord(userid string) ([]map[string]string, error) {
	sql := myorm.BSQL().Select("*").From(AWARD_TABLE_NAME).Where("user_id=? and awards_type!=0").Order("time desc").GetSQL()
	return myorm.Query(sql, userid)
}

func (this *Award) Create() {
	sql := myorm.BSQL().Insert(AWARD_CODE_TABLE_NAME).Values("code", "type", "status").GetSQL()
	for i := 0; i < 5000; i++ {
		myorm.Insert(sql, encrypt.DefaultMd5.Encode(convert.ToString(i)), 1, 0)
	}
	for i := 0; i < 800; i++ {
		myorm.Insert(sql, encrypt.DefaultMd5.Encode(convert.ToString(i)), 2, 0)
	}
	for i := 0; i < 100; i++ {
		myorm.Insert(sql, encrypt.DefaultMd5.Encode(convert.ToString(i)), 3, 0)
	}
}

// 猜字谜
func (this *Award) Word(uid string, ok bool) (int, string, error) {
	if ok {
		return this.Get(uid, map[int]int{
			1: 20,
			2: 0,
			3: 0,
		}, 1)
	} else {
		// 添加中奖记录
		tx, err := myorm.Begin()
		if err != nil {
			return 0, "", err
		}
		info := Award{}
		info.UserId = uid
		info.Time = convert.ToInt(timestamp.GetHourTimestamp(0))
		info.Date = convert.ToInt(timestamp.GetDayTimestamp(0))
		info.Code = ""
		info.AwardsType = 0
		info.Source = 1
		if this.addReocrd(tx, info) != nil {
			return 0, "", err
		}
		tx.Commit()
	}
	return 0, "", nil
}

// 获奖记录
func (this *Award) Records(source string) ([]map[string]string, error) {
	sql := "SELECT ar.awards_type,u.username,u.phone from 221su_award_record ar left JOIN 221su_users u on ar.user_id = u.id " +
		"where ar.awards_type!=0 and source=? order by ar.time desc limit 30"
	return myorm.Query(sql, source)
}

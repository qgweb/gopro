package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/bitly/go-simplejson"
	"github.com/ngaut/log"
	"github.com/qgweb/gopro/lib/http"
	"github.com/qgweb/gopro/lib/httpsqs"
	"github.com/qgweb/new/lib/config"
	"github.com/qgweb/new/lib/convert"
	"github.com/qgweb/new/lib/timestamp"
	"goclass/orm"
	"strings"
	"time"
)

type Putin struct {
}
type Order interface {
	Handler() error
	Check() error
}
type Advert struct {
	id     string
	status string
	errors []string
}
type Strategy struct {
	id     string
	status string
	errors []string
}

var (
	mysqlsession *orm.QGORM
	iniFile      config.ConfigContainer
	err          error
	info         = map[string]string{
		"1000": "没有推广计划",
		"1001": "没有投放创意",
		"1002": "审核未通过可投放创意",
		"1003": "无可投放创意",
		"1004": "上层计划未打开",
		"1005": "未通过审核",
		"2000": "余额不足",
		"2001": "日限额不足",
		"3000": "不在投放周期",
		"3001": "不在投放时段",
		"4444": "数据更新失败",
		"9999": "正在投放",
		"1111": "未投放",
	}

	table_advert         = "nxu_advert"
	table_strategy       = "nxu_strategy"
	table_report_day_tmp = "nxu_report_ad_day_tmp"
)

func (this Putin) StatusHandler(id, category, status string) []byte {
	var order Order
	if id == "" || category == "" || status == "" {
		return jsonReturn("", errors.New("参数不能为空"))
	}
	switch category {
	case "advert":
		order = &Advert{id: id, status: status}
	case "strategy":
		order = &Strategy{id: id, status: status}
	default:
		jsonReturn("", errors.New("类型错误"))
	}
	getMysqlSession()
	err := handleData(order)
	if err != nil {
		return jsonReturn("", err)
	}
	return jsonReturn("", err)
}

/**
 * 处理计划或者创意
 */
func handleData(order Order) error {
	err := order.Check()
	if err != nil {
		return err
	}
	err = order.Handler()
	if err != nil {
		return err
	}
	return nil
}

/**
 * 检查创意
 */
func (this *Advert) Check() error {
	today := timestamp.GetDayTimestamp(0)
	AdvertInfo, err := GetAdvertInfo(this.id)
	if err != nil {
		log.Error(err)
		return err
	}
	//检查用户余额
	money := convert.ToFloat64(getUserMoney(AdvertInfo["user_id"]))
	if money <= 0 {
		return errors.New("2000") //余额不足
	}
	//检查日限额
	StrategyInfo, err := GetStrategyInfo(AdvertInfo["strategy_id"])
	if err != nil {
		log.Error(err)
		return err
	}
	dayMoney := GetDayCostMoney(StrategyInfo["id"], today)
	if dayMoney >= convert.ToFloat64(StrategyInfo["day_limit"]) {
		return errors.New("2001") //日限额不足
	}
	//检查投放周期
	if today < StrategyInfo["begin_time"] || today > StrategyInfo["end_time"] {
		return errors.New("3000") //不在投放周期
	}
	//检查时间段
	t := time.Unix(convert.ToInt64(today), 0)
	week := convert.ToInt64(t.Weekday())
	hour := convert.ToInt64(t.Hour())
	interval := strings.Split(StrategyInfo["time_interval"], "|")
	if !CheckIntervel(interval, int(week), int(hour)) {
		return errors.New("3001") //当前时间段不投
	}
	//检查是否审核通过
	if !this.isReview() {
		return errors.New("1003") //审核未通过
	}
	//检查上层计划是否开启
	if StrategyInfo["status"] == "0" {
		return errors.New("1004") //上级计划未打开
	}
	return nil
}

/**
 * 检查通过后更新数据库
 * 入队列
 */
func (this *Advert) Handler() error {
	mysqlsession.BSQL().Update(table_advert).Set("status", "put_status").Where("id=" + this.id)
	fmt.Println(mysqlsession.LastSql())
	affectRows, err := mysqlsession.Update(1, 1)
	if err != nil {
		log.Error(err)
		return err
	}
	if affectRows == 0 {
		log.Error("修改", this.id, "的创意开关状态失败")
		return err
	}

	sqs, _ := GetSqs()
	queue := iniFile.String("httpsqs::queue")
	sqs.SetQueue(queue)
	params := map[string]string{
		"id":     this.id,
		"type":   "advert",
		"status": this.status,
	}
	err = sqs.Put(params)
	if err != nil {
		log.Error(err)
	}
	fmt.Println(sqs.GetLastUrl())
	return nil
}

/**
 * 检查是否审核通过
 */
func (this *Advert) isReview() bool {
	mysqlsession.BSQL().Select("id").From(table_advert).Where("review_status=1").And("id=" + this.id)
	result, err := mysqlsession.Query()
	if err != nil {
		log.Error(err)
		return false
	}
	if len(result) == 0 {
		return false
	}
	return true
}

/**
 * 获取创意信息
 */
func GetAdvertInfo(id string) (map[string]string, error) {
	mysqlsession.BSQL().Select("*").From(table_advert).Where("id =" + id)
	advert, err := mysqlsession.Query()
	if err != nil {
		log.Error(err)
		return nil, err
	}
	if len(advert) == 0 {
		return nil, errors.New("Can't find advert")
	}
	return advert[0], err
}

/**
 * 获取计划信息
 */
func GetStrategyInfo(id string) (map[string]string, error) {
	mysqlsession.BSQL().Select("*").From(table_strategy).Where("id =" + id)
	strategy, err := mysqlsession.Query()
	if err != nil {
		log.Error(err)
		return nil, err
	}
	if len(strategy) == 0 {
		return nil, errors.New("Can't find strategy")
	}
	return strategy[0], err
}

/**
 * 获取当日已消耗
 */
func GetDayCostMoney(id, date string) float64 {
	mysqlsession.BSQL().Select("SUM(money) as money").From(table_report_day_tmp).Where("strategy_id =" + id).And("date=" + date)
	result, err := mysqlsession.Query()
	if err != nil {
		log.Error(err)
		return 0
	}
	money := result[0]["money"]
	if money == "" {
		money = "0"
	}
	return convert.ToFloat64(money)
}

/**
 * 检查计划
 */
func (this *Strategy) Check() error {
	// var a = make([]map[string]string, 0)
	return nil
}

func (this *Strategy) Handler() error {
	return nil
}

/**
 * 获取mysql实例
 */
func getMysqlSession() {
	if mysqlsession != nil {
		return
	}
	mysqlsession = orm.NewORM()
	iniFile, err = config.NewConfig("ini", *conf)
	if err != nil {
		log.Error("open configfile error:", err)
		return
	}
	var (
		host    = iniFile.String("mysql-9xu::host")
		port    = iniFile.String("mysql-9xu::port")
		user    = iniFile.String("mysql-9xu::user")
		pwd     = iniFile.String("mysql-9xu::pwd")
		db      = iniFile.String("mysql-9xu::db")
		charset = "utf8"
	)
	err = mysqlsession.Open(fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s", user, pwd, host, port, db, charset))
	if err != nil {
		log.Error("db connect error:", err)
		return
	}
	mysqlsession.SetMaxIdleConns(50)
	mysqlsession.SetMaxOpenConns(50)
}

/**
 * 获取sqs实例
 */
func GetSqs() (*httpsqs.Httpsqs, error) {
	host := iniFile.String("httpsqs::host")
	port := iniFile.String("httpsqs::port")
	auth := iniFile.String("httpsqs::auth")
	sqs, err := httpsqs.New(host, port, auth)
	if err != nil {
		return nil, err
	}
	return sqs, err
}

/**
 * 获取用户余额
 */
func getUserMoney(user_id string) string {
	url := "http://login.juhuisuan.com/interface/ht"
	key := iniFile.String("mysql-9xu::key")
	field, _ := json.Marshal([]string{"money"})
	where, _ := json.Marshal(map[string]string{"id": user_id})
	param := map[string]string{
		"type":    "user",
		"_action": "qg_get",
		"_field":  string(field),
		"_where":  string(where),
	}
	result, err := http.Post(url, param, key)
	if err != nil {
		log.Fatal(err)
	}
	json, _ := simplejson.NewJson(result)
	money, _ := json.Get("data").Get("data").Get("money").String()
	return money
}

/**
 * 截取时间段
 */
func CheckIntervel(intervel []string, week, hour int) bool {
	d := intervel[week]
	h := []byte(d)
	if string(h[hour]) == "1" {
		return true
	} else {
		return false
	}
}

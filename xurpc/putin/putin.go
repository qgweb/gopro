package putin

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/bitly/go-simplejson"
	"github.com/garyburd/redigo/redis"
	"github.com/ngaut/log"
	"github.com/qgweb/gopro/lib/http"
	"github.com/qgweb/gopro/lib/httpsqs"
	"github.com/qgweb/gopro/xurpc/common"
	"github.com/qgweb/gopro/xurpc/lib"
	"github.com/qgweb/new/lib/config"
	"github.com/qgweb/new/lib/convert"
	"github.com/qgweb/new/lib/timestamp"
	"goclass/orm"
	"strings"
	"sync"
	"time"
)

type Putin struct {
}

type Order interface {
	Handler()
	Check()
	GetPutStatus() string
	SaveAdvert() error
}

type Advert struct {
	id          string
	status      string
	put_status  string
	strategy_id string
}

type Strategy struct {
	id         string
	status     string
	put_status string
	advert_ids []string //计划下的创意id
}

var (
	mysqlsession *orm.QGORM
	mux1         sync.Mutex
	iniFile      config.ConfigContainer
	err          error
	info         = map[string]string{
		"1003": "无可投放创意",
		"1004": "上层计划未打开",
		"1005": "未通过审核",
		"2000": "余额不足",
		"2001": "日限额不足",
		"3000": "不在投放周期",
		"3001": "不在投放时段",
		"9999": "正在投放",
		"1111": "未投放",
	}

	put_status = map[string]string{
		"6": "无可投放创意",
		"7": "上层计划未打开",
		"8": "未通过审核",
		"2": "余额不足",
		"3": "日限额不足",
		"4": "不在投放周期",
		"5": "不在投放时段",
		"1": "正在投放",
		"0": "未投放",
	}

	table_advert         = "nxu_advert"
	table_strategy       = "nxu_strategy"
	table_report_day_tmp = "nxu_report_ad_day_tmp"
	rc_pool              *redis.Pool
)

func init() {
	if err := initConfig(); err != nil {
		log.Fatal(err)
		return
	}
	if err := initRedisPool(); err != nil {
		log.Fatal(err)
		return
	}
}

func initConfig() error {
	iniFile, err = common.GetBeegoIniObject()
	return err
}

func initRedisPool() error {
	var (
		host = iniFile.String("redis-put::host")
		port = iniFile.String("redis-put::port")
		db   = iniFile.String("redis-put::db")
	)

	if host == "" || port == "" || db == "" {
		return errors.New("redis-put配置文件缺失")
	}

	rc_pool = lib.GetRedisPool(host, port)
	conn := rc_pool.Get()
	_, err := conn.Do("SELECT", db)
	return err
}

func (this Putin) StatusHandler(id, category, status string) []byte {
	var order Order
	if id == "" || category == "" || status == "" {
		return lib.JsonReturn("", errors.New("参数不能为空"))
	}
	switch category {
	case "advert":
		order = &Advert{id: id, status: status}
	case "strategy":
		order = &Strategy{id: id, status: status}
	default:
		lib.JsonReturn("", errors.New("类型错误"))
	}
	getMysqlSession()
	handleData(order)
	msg := put_status[order.GetPutStatus()]
	return lib.JsonReturn(msg, nil)
}

/**
 * 处理计划或者创意
 */
func handleData(order Order) {
	order.Check()
	order.SaveAdvert()
	order.Handler()
}

/**
 * 检查创意
 */
func (this *Advert) Check() {
	if this.status == "0" {
		this.put_status = "0"
		return
	}
	this.put_status = "1"
	today := timestamp.GetDayTimestamp(0)
	AdvertInfo, err := GetAdvertInfo(this.id)
	if err != nil {
		log.Error(err)
		return
	}
	this.strategy_id = AdvertInfo["strategy_id"]
	//检查用户余额
	money := convert.ToFloat64(getUserMoney(AdvertInfo["user_id"]))
	if money <= 0 {
		this.put_status = "2"
	}
	//检查日限额
	StrategyInfo, err := GetStrategyInfo(AdvertInfo["strategy_id"])
	if err != nil {
		log.Error(err)
		return
	}
	dayMoney := GetDayCostMoney(StrategyInfo["id"], today)
	if dayMoney >= convert.ToFloat64(StrategyInfo["day_limit"]) {
		this.put_status = "3"
	}
	//检查投放周期
	if today < StrategyInfo["begin_time"] || today > StrategyInfo["end_time"] {
		this.put_status = "4"
	}
	//检查时间段
	t := time.Now()
	week := t.Weekday()
	hour := t.Hour()
	interval := strings.Split(StrategyInfo["time_interval"], "|")
	if !CheckIntervel(interval, int(week), hour) {
		this.put_status = "5"
	}
	//检查是否审核通过
	if !this.isReview() {
		this.put_status = "8"
		return //审核没过直接不走下去了
	}
	//检查上层计划是否开启
	if StrategyInfo["status"] == "0" {
		this.put_status = "7"
	}
}

func (this *Advert) SaveAdvert() error {
	conn := rc_pool.Get()
	defer conn.Close()

	if this.status == "0" {
		conn.Do("DEL", "advert_info:"+this.id)
	} else {
		advertInfo, err := GetAdvertInfo(this.id)
		if err != nil {
			return err
		}
		advertjson, err := json.Marshal(advertInfo)
		if err != nil {
			return err
		}

		conn.Do("SET", "advert_info:"+this.id, advertjson)
	}
	return nil
}

/**
 * 检查通过后更新数据库
 * 入队列
 */
func (this *Advert) Handler() {
	mysqlsession.BSQL().Update(table_advert).Set("status", "put_status").Where("id=" + this.id)
	_, err := mysqlsession.Update(this.status, this.put_status)
	if err != nil {
		log.Error(err)
		return
	}
	advertinfo, err := GetAdvertInfo(this.id)
	if err != nil {
		log.Error(err)
		return
	}
	this.strategy_id = advertinfo["strategy_id"]
	if this.status == "0" { //检测本计划下是否还有创意开启，没有的话修改计划的文字状态
		if !this.hasOtherAdvert() {
			mysqlsession.BSQL().Update(table_strategy).Set("put_status").Where("id=" + this.strategy_id)
			_, err := mysqlsession.Update(6)
			if err != nil {
				log.Error(err)
				return
			}
		}
		this.PutQueue()
	} else {
		if this.isStrategyOpen() && this.put_status == "1" {
			mysqlsession.BSQL().Update(table_strategy).Set("put_status").Where("id=" + this.strategy_id)
			_, err := mysqlsession.Update(1)
			if err != nil {
				log.Error(err)
				return
			}
			this.PutQueue() //上层计划开启才会入队列
		}
	}
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
 * 入队列
 */
func (this *Advert) PutQueue() {
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
		return
	}
}

func (this *Strategy) PutQueue() {
	sqs, _ := GetSqs()
	queue := iniFile.String("httpsqs::queue")
	sqs.SetQueue(queue)
	params := map[string]string{
		"id":     this.id,
		"type":   "strategy",
		"status": this.status,
	}
	err = sqs.Put(params)
	if err != nil {
		log.Error(err)
		return
	}
}

/**
 * 检查是否还有同级的
 */
func (this *Advert) hasOtherAdvert() bool {
	mysqlsession.BSQL().Select("id").From(table_advert).Where("review_status=1").And("isdel=0").And("strategy_id=" + this.strategy_id).And("id !=" + this.id).And("status = 1")
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

func (this *Advert) isStrategyOpen() bool {
	mysqlsession.BSQL().Select("status").From(table_strategy).Where("id = " + this.strategy_id)
	result, err := mysqlsession.Query()
	if err != nil {
		log.Error(err)
		return false
	}
	status, ok := result[0]["status"]
	if !ok {
		return false
	}
	if status == "0" {
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
func (this *Strategy) Check() {
	if this.status == "0" {
		this.put_status = "0"
		return
	}
	this.put_status = "1"
	today := timestamp.GetDayTimestamp(0)
	StrategyInfo, err := GetStrategyInfo(this.id)
	if err != nil {
		log.Error(err)
		return
	}
	//检查用户余额
	money := convert.ToFloat64(getUserMoney(StrategyInfo["user_id"]))
	if money <= 0 {
		this.put_status = "2"
	}
	dayMoney := GetDayCostMoney(StrategyInfo["id"], today)
	if dayMoney >= convert.ToFloat64(StrategyInfo["day_limit"]) {
		this.put_status = "3"
	}
	//检查投放周期
	if today < StrategyInfo["begin_time"] || today > StrategyInfo["end_time"] {
		this.put_status = "4"
	}
	//检查时间段
	t := time.Now()
	week := t.Weekday()
	hour := t.Hour()
	interval := strings.Split(StrategyInfo["time_interval"], "|")
	if !CheckIntervel(interval, int(week), hour) {
		this.put_status = "5"
	}
	//检查是否有符合条件的创意
	advert_ids := this.GetChildAdvert()
	if len(advert_ids) == 0 {
		this.put_status = "6"
	} else {
		this.advert_ids = advert_ids
	}
}

func (this *Strategy) SaveAdvert() error {
	conn := rc_pool.Get()
	defer conn.Close()

	advert_ids := this.GetChildAdvert()

	if this.status == "0" {
		for _, id := range advert_ids {
			conn.Do("DEL", "advert_info:"+id)
		}
	} else {
		for _, id := range advert_ids {
			advertInfo, err := GetAdvertInfo(id)
			if err != nil {
				return err
			}
			advertjson, err := json.Marshal(advertInfo)
			if err != nil {
				return err
			}

			conn.Do("SET", "advert_info:"+id, advertjson)
		}
	}
	return nil
}

func (this *Strategy) Handler() {
	//修改自身状态
	mysqlsession.BSQL().Update(table_strategy).Set("status", "put_status").Where("id=" + this.id)
	_, err := mysqlsession.Update(this.status, this.put_status)
	if err != nil {
		log.Error(err)
		return
	}
	if this.status == "1" {
		if this.put_status == "1" {
			for _, advert := range this.advert_ids {
				mysqlsession.BSQL().Update(table_advert).Set("status", "put_status").Where("id=" + advert)
				_, err := mysqlsession.Update(this.status, this.put_status)
				if err != nil {
					log.Error(err)
					return
				}
			}
		} else {
			if this.put_status != "6" {
				for _, advert := range this.advert_ids {
					mysqlsession.BSQL().Update(table_advert).Set("put_status").Where("id=" + advert)
					_, err := mysqlsession.Update(this.put_status)
					if err != nil {
						log.Error(err)
						return
					}
				}
			}
		}
	} else { //关闭的话开启的创意状态要改成上层未开启
		this.advert_ids = this.GetChildAdvert()
		for _, advert := range this.advert_ids {
			mysqlsession.BSQL().Update(table_advert).Set("put_status").Where("status=1").And("id=" + advert)
			_, err := mysqlsession.Update(7)
			if err != nil {
				log.Error(err)
				return
			}
		}
	}

	this.PutQueue()
}

func (this *Strategy) GetChildAdvert() []string {
	advert_ids := make([]string, 0, 20)
	mysqlsession.BSQL().Select("id").From(table_advert).Where("review_status=1").And("isdel=0").And("strategy_id=" + this.id)
	result, err := mysqlsession.Query()
	if err != nil {
		log.Error(err)
		return nil
	}
	for _, value := range result {
		advert_ids = append(advert_ids, value["id"])
	}
	return advert_ids
}

func (this *Strategy) GetPutStatus() string {
	return this.put_status
}

func (this *Advert) GetPutStatus() string {
	return this.put_status
}

/**
 * 获取mysql实例
 */
func getMysqlSession() {
	mux1.Lock()
	defer mux1.Unlock()
	if mysqlsession == nil {
		mysqlsession = orm.NewORM()

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

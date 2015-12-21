package lonlat

import (
	"errors"
	"fmt"
	"github.com/ngaut/log"
	"github.com/qgweb/gopro/lib/convert"
	"github.com/qgweb/gopro/lib/orm"
	"io/ioutil"
	"runtime/debug"

	"github.com/codegangsta/cli"
	"github.com/qgweb/gopro/9xutool/common"
	"gopkg.in/ini.v1"
	"gopkg.in/mgo.v2/bson"
)

type (
	UserCdTrace struct {
		mp          *common.MgoPool
		mysql       *orm.QGORM
		iniFile     *ini.File
		debug       int
		taocat_list map[string]*TaoCat //标签分类总表 map[cid]TaoCat
		tags_num    map[string]int     //标签计数
		tagsByJwd   map[string]*TagInfo
	}

	TaoCat struct { //数据模型
		Name  string
		Level int
		Cid   string
		Pid   string
	}
)

const ( //表名
	TAOCAT_TABLE     string = "taocat"
	JWD_TABLE        string = "tbl_map"
	USERACTION_TABLE string = "useraction"
)

func NewUserCdCli() cli.Command {
	return cli.Command{
		Name:  "tags_num_by_coordinate",
		Usage: "根据经纬度和ad汇总标签昨日总数",
		Action: func(c *cli.Context) {
			defer func() {
				if msg := recover(); msg != nil {
					log.Error(msg)
					debug.PrintStack()
				}
			}()

			// 获取配置文件
			filePath := common.GetBasePath() + "/conf/jw.conf"
			f, err := ioutil.ReadFile(filePath)
			if err != nil {
				log.Fatal(err)
				debug.PrintStack()
				return
			}

			ur := &UserCdTrace{}
			ur.iniFile, _ = ini.Load(f)

			// mgo 配置文件
			mconf := &common.MgoConfig{}
			mconf.DBName = ur.iniFile.Section("mongo-data_source").Key("db").String()
			mconf.Host = ur.iniFile.Section("mongo-data_source").Key("host").String()
			mconf.Port = ur.iniFile.Section("mongo-data_source").Key("port").String()
			mconf.UserName = ur.iniFile.Section("mongo-data_source").Key("user").String()
			mconf.UserPwd = ur.iniFile.Section("mongo-data_source").Key("pwd").String()
			ur.debug, _ = ur.iniFile.Section("mongo-data_source").Key("debug").Int()
			ur.mp = common.NewMgoPool(mconf)

			//mysql 配置文件
			ur.mysql = orm.NewORM()

			ur.initData()
			ur.Do(c)
		},
	}
}

func (this *UserCdTrace) initData() {
	db := this.iniFile.Section("mongo-data_source").Key("db").String()
	sess := this.mp.Get()
	defer sess.Close()

	var list []map[string]interface{}
	err := sess.DB(db).C(TAOCAT_TABLE).Find(bson.M{"type": "0"}).All(&list)
	if err != nil {
		log.Error(err)
	}

	this.taocat_list = make(map[string]*TaoCat)
	if len(list) > 0 {
		for _, v := range list {
			//处理等级map
			category := &TaoCat{
				Name:  v["name"].(string),
				Level: v["level"].(int),
				Cid:   v["cid"].(string),
				Pid:   v["pid"].(string),
			}
			this.taocat_list[category.Cid] = category
		}
	}
	log.Info("初始化taocat_list完毕")
}

//整体逻辑
//从经纬度文件中提取出ad
//for执行getTagsInfo方法，处理标签
//最后把tags_num入库，入库的时候再做映射取标签中文名
func (this *UserCdTrace) Do(c *cli.Context) {
	var (
		db   = this.iniFile.Section("mongo-lonlat_data").Key("db").String()
		sess = this.mp.Get()
	)
	iter := sess.DB(db).C(JWD_TABLE).Find(nil).Iter()
	i := 1
	for {
		var info map[string]interface{}
		if !iter.Next(&info) {
			break
		}
		ad := info["AD"].(string)
		this.getTagsInfo(ad)
		if this.debug == 1 {
			log.Info("已处理", convert.ToString(i), "条记录")
		}
		i++
	}

	if len(this.tags_num) > 0 {
		this.getMysqlConnect() //连接mysql

		DayTimestamp := common.GetDayTimestamp(-1)
		for tag_id, num := range this.tags_num {
			if _, ok := this.taocat_list[tag_id]; !ok {
				continue
			}
			tag_text := this.taocat_list[tag_id].Name
			this.mysql.BSQL().Insert("tags_report_day").Values("tag_id", "tag_text", "num", "time")
			_, err := this.mysql.Insert(tag_id, tag_text, num, DayTimestamp)
			if err != nil {
				log.Warn("插入失败 ", err)
			}
		}
	}
	log.Info("数据分析完毕!")
}

func (this *UserCdTrace) getMysqlConnect() {
	var (
		db      = this.iniFile.Section("mysql-jw").Key("db").String()
		host    = this.iniFile.Section("mysql-jw").Key("host").String()
		port    = this.iniFile.Section("mysql-jw").Key("port").String()
		user    = this.iniFile.Section("mysql-jw").Key("user").String()
		pwd     = this.iniFile.Section("mysql-jw").Key("pwd").String()
		charset = "utf8"
		err     = this.mysql.Open(fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s",
			user, pwd, host, port, db, charset))
	)
	if err != nil {
		log.Fatal(err)
	}
	this.mysql.SetMaxIdleConns(50)
	this.mysql.SetMaxOpenConns(100)
}

//根据ad获取标签
func (this *UserCdTrace) getTagsInfo(ad string) {
	var (
		db        = this.iniFile.Section("mongo-data_source").Key("db").String()
		sess      = this.mp.Get()
		timestamp = common.GetDayTimestamp(-1) //0为今日
		err       error
	)
	defer sess.Close()

	iter := sess.DB(db).C(USERACTION_TABLE).Find(bson.M{"AD": ad, "timestamp": timestamp}).Iter()

	this.tags_num = make(map[string]int)
	for {
		var info map[string]interface{}
		if !iter.Next(&info) {
			break
		}
		for _, tag := range info["tag"].([]interface{}) { //获取每个ad内的tagid
			tagm := tag.(map[string]interface{})
			if tagm["tagmongo"].(string) == "1" { //如果是mongoid忽略
				continue
			}
			cid := tagm["tagId"].(string)
			cg := this.taocat_list[cid] //从总标签的map判断是否是3级标签
			if cg.Level != 3 {
				cid, err = cg.getLv3Id(this)
				if err != nil {
					continue
				}
			}
			this.tags_num[cid] = this.tags_num[cid] + 1
		}
	}
}

//获取相应的三级标签id
func (this *TaoCat) getLv3Id(uc *UserCdTrace) (string, error) {
	if this.Level == 3 {
		return this.Cid, nil
	}
	n := this.Level - 3
	if n < 0 { //如果是
		return "", errors.New("标签等级过高")
	}
	tmp := *this
	for i := 0; i < n; i++ {
		tmp = *uc.taocat_list[tmp.Pid] //获取父级
	}
	return tmp.Cid, nil
}

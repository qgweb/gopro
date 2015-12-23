package lonlat

import (
	"errors"
	"fmt"
	"github.com/ngaut/log"
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
		mp_jw       *common.MgoPool
		mysql       *orm.QGORM
		iniFile     *ini.File
		debug       int
		debug_jw    int
		taocat_list map[string]*TaoCat //标签分类总表 map[cid]TaoCat
		tags_num    map[string]int     //标签计数
		tagsByJwd   map[string]*TagInfo
		uniqueUser  map[string]int //用户去重 md5(ad+ua+cid)
	}

	TaoCat struct { //数据模型
		Name  string
		Level int
		Cid   string
		Pid   string
	}

	TagInfo struct { //数据模型
		tagid string
		lon   string
		lat   string
		num   int
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

			// mgo 配置文件 经纬
			mconf_jw := &common.MgoConfig{}
			mconf_jw.DBName = ur.iniFile.Section("mongo-lonlat_data").Key("db").String()
			mconf_jw.Host = ur.iniFile.Section("mongo-lonlat_data").Key("host").String()
			mconf_jw.Port = ur.iniFile.Section("mongo-lonlat_data").Key("port").String()
			mconf_jw.UserName = ur.iniFile.Section("mongo-lonlat_data").Key("user").String()
			mconf_jw.UserPwd = ur.iniFile.Section("mongo-lonlat_data").Key("pwd").String()
			ur.debug_jw, _ = ur.iniFile.Section("mongo-lonlat_data").Key("debug").Int()
			ur.mp_jw = common.NewMgoPool(mconf_jw)

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

func (this *UserCdTrace) Do(c *cli.Context) {
	var (
		db    = this.iniFile.Section("mongo-data_source").Key("db").String()
		sess  = this.mp.Get()
		begin = common.GetDayTimestamp(-1)
		end   = common.GetDayTimestamp(0)
	)
	this.tags_num = make(map[string]int) //最终数据保存
	// iter := sess.DB(db).C(USERACTION_TABLE).Find(bson.M{"timestamp": "1449417600"}).Limit(5).Iter()
	iter := sess.DB(db).C(USERACTION_TABLE).Find(bson.M{"timestamp": bson.M{"$gt": begin, "$lte": end}}).Iter()
	i := 1
	for {
		var userInfo map[string]interface{}
		if !iter.Next(&userInfo) {
			break
		}
		this.getTagsInfo(userInfo)
		if this.debug_jw == 1 {
			log.Info("已处理", i, "条记录")
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
func (this *UserCdTrace) getTagsInfo(userInfo map[string]interface{}) {
	var (
		db   = this.iniFile.Section("mongo-lonlat_data").Key("db").String()
		sess = this.mp_jw.Get()
	)
	defer sess.Close()

	if ad, ok := userInfo["AD"]; ok { //userInfo是useraction取出来的数据
		has, err := sess.DB(db).C(JWD_TABLE).Find(bson.M{"ad": ad.(string)}).Count() //从经纬度判断是否有这个用户
		if err != nil {
			log.Info(err)
		}

		if has > 0 {
			for _, tag := range userInfo["tag"].([]interface{}) {
				tagm := tag.(map[string]interface{})

				if tagm["tagmongo"].(string) == "1" { //如果是mongoid忽略
					continue
				}
				cid := tagm["tagId"].(string)
				if _, ok := this.taocat_list[cid]; !ok { //从总标签的map判断是否是3级标签
					log.Info(cid)
					continue
				}
				cg := this.taocat_list[cid]
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

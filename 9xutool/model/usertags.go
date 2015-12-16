package model

import (
	"github.com/qgweb/gopro/lib/convert"
	"github.com/qgweb/gopro/lib/encrypt"
	"github.com/qgweb/gopro/lib/grab"
	"gopkg.in/mgo.v2"
	"io/ioutil"
	"runtime"
	"runtime/debug"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/codegangsta/cli"
	"github.com/ngaut/log"
	"github.com/qgweb/gopro/9xutool/common"
	"gopkg.in/ini.v1"
	"gopkg.in/mgo.v2/bson"
)

type WaitGroup struct {
	sync.WaitGroup
}

func (this *WaitGroup) Run(fun func(...interface{}), param ...interface{}) {
	this.Add(1)
	go func() {
		fun(param...)
		this.Done()
	}()
}

type TAGSTrace struct {
	mp                  *common.MgoPool
	iniFile             *ini.File
	usertracelist       map[string]map[string]int
	rw                  sync.RWMutex
	totaldeal           int64
	category_tao_list   map[string]string
	category_other_list map[string]string
}

func NewTAGSTraceCli() cli.Command {
	return cli.Command{
		Name:  "get_tags_by_cookie",
		Usage: "根据用户ua和ad汇总十天内的标签",
		Action: func(c *cli.Context) {
			defer func() {
				if msg := recover(); msg != nil {
					log.Error(msg)
				}
			}()

			// 获取配置文件
			filePath := common.GetBasePath() + "/conf/ut.conf"
			f, err := ioutil.ReadFile(filePath)
			if err != nil {
				log.Fatal(err)
				debug.PrintStack()
				return
			}

			ur := &TAGSTrace{}
			ur.iniFile, _ = ini.Load(f)
			ur.usertracelist = make(map[string]map[string]int)
			ur.category_other_list = make(map[string]string)
			ur.category_tao_list = make(map[string]string)

			// mgo 配置文件
			mconf := &common.MgoConfig{}
			mconf.DBName = ur.iniFile.Section("mongo").Key("db").String()
			mconf.Host = ur.iniFile.Section("mongo").Key("host").String()
			mconf.Port = ur.iniFile.Section("mongo").Key("port").String()
			mconf.UserName = ur.iniFile.Section("mongo").Key("user").String()
			mconf.UserPwd = ur.iniFile.Section("mongo").Key("pwd").String()
			ur.mp = common.NewMgoPool(mconf)

			ur.initData()
			ur.Do(c)
		},
	}
}

func (this *TAGSTrace) initData() {
	db := this.iniFile.Section("mongo").Key("db").String()
	sess := this.mp.Get()
	defer sess.Close()

	var list []map[string]interface{}
	err := sess.DB(db).C("taocat").Find(bson.M{"type": "0"}).All(&list)
	if err != nil {
		log.Error(err)
	}

	if len(list) > 0 {
		for _, v := range list {
			this.category_tao_list[v["cid"].(string)] = v["name"].(string)
		}
	}

	var list1 []map[string]interface{}
	err = sess.DB(db).C("category").Find(nil).All(&list1)
	if err != nil {
		log.Error(err)
	}

	if len(list1) > 0 {
		for _, v := range list1 {
			this.category_other_list[v["_id"].(bson.ObjectId).Hex()] = v["name"].(string)
		}
	}

	log.Error(this.category_tao_list)
	log.Error(this.category_other_list)
}

func (this *TAGSTrace) AppendData(info map[string]interface{}) {
	UA := "ua" //可能会没有UA
	if u, ok := info["UA"]; ok {
		UA = u.(string)
	}
	key := info["AD"].(string) + "_" + UA
	for _, tag := range info["tag"].([]interface{}) {
		tagm := tag.(map[string]interface{})
		this.rw.RLock()
		if _, ok := this.usertracelist[key]; ok { //如果已经有这个用户相关的tag
			if _, ok := this.usertracelist[key][tagm["tagId"].(string)]; ok { //去重,判断是否已存在
				this.rw.RUnlock()
				continue
			} else {
				this.rw.RUnlock()
				this.rw.Lock()
				this.usertracelist[key][tagm["tagId"].(string)] = convert.ToInt(info["timestamp"].(string))
				this.rw.Unlock()
			}
		} else {
			this.rw.RUnlock()
			this.rw.Lock()
			this.usertracelist[key] = map[string]int{tagm["tagId"].(string): convert.ToInt(info["timestamp"].(string))}
			this.rw.Unlock()
		}
	}
}

func (this *TAGSTrace) ReadData(qi ...interface{}) {
	var (
		db    = this.iniFile.Section("mongo").Key("db").String()
		sess  = this.mp.Get()
		table = "useraction"
		query = qi[0].(bson.M)
	)
	defer sess.Close()

	iter := sess.DB(db).C(table).Find(query).Iter()
	for {
		var info map[string]interface{}
		if !iter.Next(&info) {
			break
		}
		this.totaldeal = atomic.AddInt64(&this.totaldeal, 1)
		if this.totaldeal%10000 == 0 {
			log.Info(this.totaldeal)
		}

		this.AppendData(info)
	}
}

func (this *TAGSTrace) Do(c *cli.Context) {
	var (
		db        = this.iniFile.Section("mongo").Key("db").String()
		sess      = this.mp.Get()
		table_put = "useraction_temp_tags"

		wg = WaitGroup{}
	)
	defer sess.Close()
	runtime.GOMAXPROCS(5)

	//检查是否有数据,有就先清空
	sess.DB(db).C(table_put).DropCollection()
	sess.DB(db).C(table_put).Create(&mgo.CollectionInfo{})
	sess.DB(db).C(table_put).EnsureIndexKey("adua")

	//查询数据先
	wg.Run(this.ReadData, bson.M{"timestamp": bson.M{"$gte": this.getDayTime(-10), "$lt": this.getDayTime(-8)}})
	wg.Run(this.ReadData, bson.M{"timestamp": bson.M{"$gte": this.getDayTime(-8), "$lt": this.getDayTime(-6)}})
	wg.Run(this.ReadData, bson.M{"timestamp": bson.M{"$gte": this.getDayTime(-6), "$lt": this.getDayTime(-4)}})
	wg.Run(this.ReadData, bson.M{"timestamp": bson.M{"$gte": this.getDayTime(-4), "$lt": this.getDayTime(-2)}})
	wg.Run(this.ReadData, bson.M{"timestamp": bson.M{"$gte": this.getDayTime(-2)}})
	wg.Wait()

	log.Warn(len(this.usertracelist))
	var (
		size     = 10000
		list_num = len(this.usertracelist)
		list_put = make([]interface{}, 0, size)
	)

	for key, value := range this.usertracelist {
		tags_data := make([]string, 0, 5)
		adua := strings.Split(key, "_")
		s := grab.NewMapSorter(value)
		edrange := len(s)
		s.Sort() //排序

		if edrange >= 6 { //取前五个
			edrange = 5
		}
		//查询出标签的中文名
		for _, r := range s[0:edrange] {
			if bson.IsObjectIdHex(r.Key) {
				if v, ok := this.category_other_list[r.Key]; ok {
					tags_data = append(tags_data, v)
				}
			} else {
				if v, ok := this.category_tao_list[r.Key]; ok {
					tags_data = append(tags_data, v)
				}
			}
		}

		list_put = append(list_put, bson.M{
			"ad":   adua[0],
			"ua":   adua[1],
			"adua": encrypt.DefaultMd5.Encode(adua[0] + adua[1]),
			"tag":  tags_data,
		})

		if len(list_put) == size || len(list_put) == list_num {
			sess.DB(db).C(table_put).Insert(list_put...)
			log.Warn(len(list_put))
			list_put = make([]interface{}, 0, size)
			list_num = list_num - size
		}
	}
	log.Info("ok")
}

/**
 * 获取十天前零点时间戳
 */
func (this *TAGSTrace) getDayTime(day int) string {
	d := time.Now().AddDate(0, 0, day).Format("20060102")
	a, _ := time.ParseInLocation("20060102", d, time.Local)
	return convert.ToString(a.Unix())
}

package model

import (
	"io/ioutil"
	"math"
	"runtime"
	"sync"
	"time"

	"github.com/codegangsta/cli"
	"github.com/ngaut/log"
	"github.com/qgweb/gopro/9xutool/common.go"
	"github.com/qgweb/gopro/lib/convert"
	"gopkg.in/ini.v1"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type UserTrace struct {
	mp      *common.MgoPool
	iniFile *ini.File
}

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

func NewUserTraceCli() cli.Command {
	return cli.Command{
		Name:  "user_trace_merge",
		Usage: "生成用户最近3天浏览轨迹,供九旭精准投放",
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
				return
			}

			ur := &UserTrace{}
			ur.iniFile, _ = ini.Load(f)

			// mgo 配置文件
			mconf := &common.MgoConfig{}
			mconf.DBName = ur.iniFile.Section("mongo").Key("db").String()
			mconf.Host = ur.iniFile.Section("mongo").Key("host").String()
			mconf.Port = ur.iniFile.Section("mongo").Key("port").String()
			mconf.UserName = ur.iniFile.Section("mongo").Key("user").String()
			mconf.UserPwd = ur.iniFile.Section("mongo").Key("pwd").String()
			ur.mp = common.NewMgoPool(mconf)

			ur.Do(c)
		},
		//		Flags: []cli.Flag{
		//			cli.StringFlag{"port, p", "3000", "Temporary port number to prevent conflict", ""},
		//			cli.StringFlag{"config, c", "custom/conf/app.ini", "Custom configuration file path", ""},
		//		},
	}
}

func (this *UserTrace) Do(c *cli.Context) {
	runtime.GOMAXPROCS(8)
	var (
		date          = time.Now()
		day           = date.Format("20060102")
		hour          = convert.ToString(date.Hour() - 1)
		b1day         = date.AddDate(0, 0, -1).Format("20060102")  //1天前
		b2day         = date.AddDate(0, 0, -2).Format("20060102")  //2天前
		b3day         = date.AddDate(0, 0, -3).Format("20060102")  //3天前
		b14day        = date.AddDate(0, 0, -14).Format("20060102") //14天前
		b15day        = date.AddDate(0, 0, -15).Format("20060102") //15天前
		db            = this.iniFile.Section("mongo").Key("db").String()
		table         = "useraction"
		table_put     = "useraction_put"
		table_put_big = "useraction_put_big"
		list          map[string][]map[string]interface{}
		mux           sync.Mutex
		taoCategory   map[string]string
	)

	//初始化淘宝分类
	taoCategory = this.getBigCat()

	list = make(map[string][]map[string]interface{})

	var appendFun = func(l []map[string]interface{}) {
		for _, v := range l {
			key := v["AD"].(string)
			if tag, ok := list[key]; ok {
				//去重
				for _, tv := range v["tag"].([]interface{}) {
					isee := false
					tvm := tv.(map[string]interface{})
					for _, tv1 := range tag {
						if tvm["tagId"] == tv1["tagId"] {
							isee = true
							break
						}
					}
					if !isee {
						mux.Lock()
						list[key] = append(list[key], tvm)
						mux.Unlock()
					}
				}
			} else {
				tag := make([]map[string]interface{}, 0, len(v["tag"].([]interface{})))
				for _, vv := range v["tag"].([]interface{}) {
					tag = append(tag, vv.(map[string]interface{}))
				}
				mux.Lock()
				list[key] = tag
				mux.Unlock()
			}
		}

	}

	// 读取数据函数
	readDataFun := func(query ...interface{}) {
		var (
			count     = 0
			page      = 1
			pageSize  = 100000
			totalPage = 0
			sess      = this.mp.Get()
			querym    bson.M
		)

		if v, ok := query[0].(bson.M); ok {
			querym = v
		} else {
			return
		}

		count, err := sess.DB(db).C(table).Find(querym).Count()
		if err != nil {
			log.Error(err)
			return
		}

		totalPage = int(math.Ceil(float64(count) / float64(pageSize)))

		for ; page <= totalPage; page++ {
			var tmpList []map[string]interface{}
			if err := sess.DB(db).C(table).Find(querym).
				Select(bson.M{"_id": 0, "AD": 1, "tag": 1}).
				Limit(pageSize).
				Skip((page - 1) * pageSize).All(&tmpList); err != nil {
				log.Error(err)
				continue
			}

			appendFun(tmpList)
			log.Warn(len(list))
		}

		sess.Close()
	}

	wg := &WaitGroup{}

	// domainId 0 电商  1 医疗 4 金融
	// 当天前一个小时前的数据
	wg.Run(readDataFun, bson.M{"day": day, "hour": bson.M{
		"$lte": hour, "$gte": "00"}, "domainId": "0",
	})

	// == 医疗金融
	wg.Run(readDataFun, bson.M{"day": day, "hour": bson.M{
		"$lte": hour, "$gte": "00"}, "domainId": bson.M{"$ne": "0"},
	})

	// 前2天数据
	wg.Run(readDataFun, bson.M{"day": bson.M{
		"$lte": b1day, "$gte": b2day}, "domainId": "0",
	})

	wg.Run(readDataFun, bson.M{"day": bson.M{
		"$lte": b1day, "$gte": b14day}, "domainId": bson.M{"$ne": "0"},
	})

	// 第前3天的小时数据
	wg.Run(readDataFun, bson.M{"day": b3day, "hour": bson.M{
		"$gte": hour, "$lte": "23"}, "domainId": "0",
	})

	wg.Run(readDataFun, bson.M{"day": b15day, "hour": bson.M{
		"$gte": hour, "$lte": "23"}, "domainId": bson.M{"$ne": "0"},
	})

	wg.Wait()

	//更新投放表
	log.Info(len(list))
	sess := this.mp.Get()
	sess.DB(db).C(table_put).DropCollection()
	sess.DB(db).C(table_put_big).DropCollection()

	//加索引
	sess.DB(db).C(table_put).Create(&mgo.CollectionInfo{})
	sess.DB(db).C(table_put_big).Create(&mgo.CollectionInfo{})
	sess.DB(db).C(table_put).EnsureIndexKey("tag.tagId")
	sess.DB(db).C(table_put_big).EnsureIndexKey("tag.tagId")

	var (
		size     = 10000
		list_num = len(list)
	)

	list_put := make([]interface{}, 0, size)
	list_put_big := make([]interface{}, 0, size)

	for k, v := range list {
		list_put = append(list_put, bson.M{
			"AD":  k,
			"tag": v,
		})

		bv := this.copy(v)
		for k, vv := range bv {
			if bv[k]["tagmongo"].(string) == "0" {
				if cid, ok := taoCategory[vv["tagId"].(string)]; ok {
					bv[k]["tagId"] = cid
				}
			}
		}

		list_put_big = append(list_put_big, bson.M{
			"AD":  k,
			"tag": bv,
		})

		if len(list_put) == size || len(list_put) == list_num {
			sess := this.mp.Get()
			sess.DB(db).C(table_put).Insert(list_put...)
			sess.DB(db).C(table_put_big).Insert(list_put_big...)
			sess.Close()

			list_put = make([]interface{}, 0, size)
			list_put_big = make([]interface{}, 0, size)
			list_num = list_num - size
			log.Warn(len(list_put))
		}
	}

	wg.Wait()

	sess.Close()
	log.Info("ok")
}

// 复制对象
func (this *UserTrace) copy(src []map[string]interface{}) []map[string]interface{} {
	var dis = make([]map[string]interface{}, 0, len(src))
	for _, v := range src {
		var node = make(map[string]interface{})
		for kk, vv := range v {
			node[kk] = vv
		}
		dis = append(dis, node)
	}

	return dis
}

// 获取大分类
func (this *UserTrace) getBigCat() map[string]string {
	var (
		db    = this.iniFile.Section("mongo").Key("db").String()
		table = "taocat"
		sess  = this.mp.Get()
	)

	defer sess.Close()

	var info []map[string]interface{}
	var list = make(map[string]string)
	err := sess.DB(db).C(table).Find(bson.M{"type": "0"}).Select(bson.M{"bid": 1, "cid": 1}).All(&info)
	if err == nil {
		for _, v := range info {
			list[v["cid"].(string)] = v["bid"].(string)
		}
		return list
	}
	return nil
}

package lonlat

import (
	"bufio"
	"fmt"
	"github.com/ngaut/log"
	"io"
	"os"
	"runtime/debug"
	"strings"

	"github.com/codegangsta/cli"
	"github.com/qgweb/gopro/9xutool/common"
	"gopkg.in/ini.v1"
	// "gopkg.in/mgo.v2/bson"
)

type LonLat struct {
	iniFile  *ini.File
	filename string //需要读取的文件名字
	mp       *common.MgoPool
	data     []*LonLatData
	debug    int //debug模式
}

type LonLatData struct {
	AD       string
	Lon, Lat string
}

func NewLonLatCli() cli.Command {
	return cli.Command{
		Name:  "getlonlatbyfile",
		Usage: "从文件中读取ad和经纬，入mongo库",
		Action: func(c *cli.Context) {
			defer func() {
				if msg := recover(); msg != nil {
					log.Error(msg)
					debug.PrintStack()
				}
			}()

			if len(c.Args()) < 1 {
				fmt.Println("[错误]请输入需要读取的经纬度文件路径")
				os.Exit(1)
			}
			_file := c.Args()[0]
			// 获取配置文件
			var err error
			filePath := common.GetBasePath() + "/conf/jw.conf"
			ll := &LonLat{}
			ll.filename = _file
			ll.iniFile, err = ini.Load(filePath)
			if err != nil {
				log.Fatal(err)
				debug.PrintStack()
				return
			}
			// mgo 配置文件
			mconf := &common.MgoConfig{}
			mconf.DBName = ll.iniFile.Section("mongo").Key("db").String()
			mconf.Host = ll.iniFile.Section("mongo").Key("host").String()
			mconf.Port = ll.iniFile.Section("mongo").Key("port").String()
			mconf.UserName = ll.iniFile.Section("mongo").Key("user").String()
			mconf.UserPwd = ll.iniFile.Section("mongo").Key("pwd").String()
			ll.debug, _ = ll.iniFile.Section("mongo").Key("debug").Int()
			ll.mp = common.NewMgoPool(mconf)

			ll.Do(c)
		},
	}
}

func (this *LonLat) Do(c *cli.Context) {
	var (
		db   = this.iniFile.Section("mongo").Key("db").String()
		sess = this.mp.Get()
		err  error
	)
	defer sess.Close()

	fmt.Println("开始读取文件，整理经纬度数据...")
	this.extData()

	fmt.Println("开始存入数据库")
	for _, v := range this.data {
		err = sess.DB(db).C("tbl_map").Insert(v)
		if err != nil {
			panic(err)
		}
	}
	fmt.Println("执行完毕!")
}

/**
 * 从文件中读取数据，放入mdata等待入库
 */
func (this *LonLat) extData() {
	f, err := os.Open(this.filename)
	if err != nil {
		log.Fatal("读取文件失败:", err)
	}
	defer f.Close()
	reader := bufio.NewReader(f)
	result := []*LonLatData{}
	i := 1
	for {
		str, err := reader.ReadString('\n')
		if err != nil && err == io.EOF {
			break
		}
		if err != nil {
			continue
		}
		_tmp := strings.Split(str, "|")
		if len(_tmp) < 2 { //如果数据格式不正确 忽略
			continue
		}
		lonlat := strings.Split(strings.TrimSpace(_tmp[1]), ",")
		if len(lonlat) < 2 {
			continue
		}
		lonlatdata := &LonLatData{
			AD:  _tmp[0],
			Lon: lonlat[0],
			Lat: lonlat[1],
		}
		result = append(result, lonlatdata)
		if this.debug == 1 {
			fmt.Println("已读取", i, "条经纬度数据")
		}
		i++
	}
	this.data = result
	fmt.Println("整理经纬度完毕")
}
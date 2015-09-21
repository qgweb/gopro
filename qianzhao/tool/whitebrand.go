package main

import (
	"bufio"
	"flag"
	"github.com/qgweb/gopro/lib/convert"
	"github.com/qgweb/gopro/qianzhao/model"
	"github.com/qiniu/iconv"
	"io"
	"log"
	"os"
	"strings"
	"time"
)

var (
	file   = flag.String("file", "", "数据文件")
	path   = flag.String("path", "", "数据文件路径")
	prefix = flag.String("prefix", "radius_gongxin_quansheng_school_user", "数据文件前缀")
)

func init() {
	flag.Parse()

	if *file == "" && *path == "" {
		log.Fatalln("数据文件参数不存在")
	}

	if *file == "" {
		*file = *path + "/" + *prefix + time.Now().Add(-time.Hour*24).Format("20060102.txt")
	}
}

func main() {
	f, err := os.Open(*file)
	if err != nil {
		log.Fatalln("文件打开失败：", err)
	}
	defer f.Close()
	cd, _ := iconv.Open("utf-8", "gbk")
	bi := bufio.NewReader(f)
	for {
		line, err := bi.ReadString('\n')
		if err == io.EOF {
			break
		}

		line = strings.TrimSpace(line)
		line = cd.ConvString(line)
		//宽带账户|属地|校园名称|校园组别|上行带宽|下行带宽
		datas := strings.Split(line, "^$^")
		if len(datas) < 6 {
			log.Println("数据出错：", line)
			continue
		}
		log.Println(datas)

		ba := model.BrandAccount{}
		ba.Account = datas[0]
		ba.Area = datas[1]
		ba.SchoolName = datas[3]
		ba.SchoolGroup = datas[4]
		ba.UpBroadband = convert.ToInt(datas[5])
		ba.DownBroadband = convert.ToInt(datas[6])
		ba.TotalTime = 3600
		ba.UsedTime = 0
		ba.TryCount = 0

		if info, _ := ba.GetAccountInfo(ba.Account); info.Id != "" {
			ba.UsedTime = info.UsedTime
			ba.TryCount = info.TryCount
			ba.EditAccount(ba)
		} else {
			ba.AddBroadBand(ba)
		}
	}
}

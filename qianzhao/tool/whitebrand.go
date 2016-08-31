package main

import (
	"bufio"
	"flag"
	"github.com/qgweb/gopro/lib/convert"
	"github.com/qgweb/gopro/qianzhao/model"
	"io"
	"github.com/ngaut/log"
	"os"
	"strings"
	"time"
	"golang.org/x/net/html/charset"
	"io/ioutil"
)

var (
	file = flag.String("file", "", "数据文件")
	path = flag.String("path", "", "数据文件路径")
	prefix = flag.String("prefix", "radius_gongxin_quansheng_school_user", "数据文件前缀")
)

func changeCharsetEncodingAuto(sor io.ReadCloser) string {
	var err error
	destReader, err := charset.NewReader(sor, "text/html; charset=gbk")

	if err != nil {
		log.Error(err)
		destReader = sor
	}

	var sorbody []byte
	if sorbody, err = ioutil.ReadAll(destReader); err != nil {
		log.Error(err)
	}

	bodystr := string(sorbody)

	return bodystr
}

func init() {
	flag.Parse()

	if *file == "" && *path == "" {
		log.Fatal("数据文件参数不存在")
	}

	if *file == "" {
		*file = *path + "/" + *prefix + time.Now().Add(-time.Hour * 24).Format("20060102.txt")
	}
}

func main() {
	f, err := os.Open(*file)
	if err != nil {
		log.Fatal("文件打开失败：", err)
	}
	defer f.Close()
	bi := bufio.NewReader(f)
	for {
		line, err := bi.ReadString('\n')
		if err == io.EOF {
			break
		}

		line = strings.TrimSpace(line)
		//line = cd.ConvString(line)
		//宽带账户|属地|校园名称|校园组别|上行带宽|下行带宽
		log.Error(line)
		line = changeCharsetEncodingAuto(ioutil.NopCloser(strings.NewReader(line)))
		datas := strings.Split(line, "^$^")
		if len(datas) < 6 {
			log.Error("数据出错：", line)
			continue
		}
		log.Info(datas)

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

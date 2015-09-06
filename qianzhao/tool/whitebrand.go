package main

import (
	"bufio"
	"flag"
	"github.com/qgweb/gopro/qianzhao/model"
	"io"
	"log"
	"os"
	"strings"
)

var (
	file = flag.String("file", "", "数据文件")
)

func init() {
	flag.Parse()

	if *file == "" {
		log.Fatalln("数据文件参数不存在")
	}
}

func main() {
	f, err := os.Open(*file)
	if err != nil {
		log.Fatalln("文件打开失败：", err)
	}
	defer f.Close()

	bi := bufio.NewReader(f)
	for {
		line, err := bi.ReadString('\n')
		if err == io.EOF {
			break
		}

		line = strings.TrimSpace(line)
		//宽带账户|属地|校园名称|校园组别|上行带宽|下行带宽
		datas := strings.Split(line, "|")
		if len(datas) != 6 {
			log.Println("数据出错："，  line)
			continue
		}

		ba := model.BrandAccount{}
		ba.Account = datas[0]
		ba.Area = datas[1]
		ba.SchoolName = datas[2]
		ba.SchoolGroup = datas[3]
		ba.UpBroadband = datas[4]
		ba.DownBroadband = datas[5]

		if ba.AccountExist(ba) {
			ba.EditAccount(ba)
		} else {
			ba.AddBroadBand(ba)
		}
	}
}

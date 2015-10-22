// 异常停止，处理中断用户
package server

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/qgweb/gopro/lib/convert"
	"github.com/qgweb/gopro/lib/grab"
	"github.com/qgweb/gopro/qianzhao-dz-tcpserver/common/function"
	"github.com/qgweb/gopro/qianzhao-dz-tcpserver/model"
)

func GetLastFile(dir string) []byte {
	fileNames := make(map[string]int)
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		fileNames[path] = int(info.ModTime().Unix())
		return nil
	})

	m := grab.NewMapSorter(fileNames)
	m.Sort()
	if len(m) > 0 {
		n := m[0].Key
		d, err := ioutil.ReadFile(n)
		if err != nil {
			log.Println("打开数据文件失败,错误信息为:", err)
			return nil
		}
		return d
	}
	return nil
}

func DealData(data []byte) {
	list := make(map[string]*Account)
	err := json.Unmarshal(data, &list)
	if err != nil {
		log.Println("反序列化数据文件失败")
		return
	}

	if len(list) > 0 {
		bd := &BDInterfaceManager{}
		for k, v := range list {
			log.Println(k, v)
			// 调用停止接口
			bd.Stop(v.ChannelId)

			// 把记录添加到数据库
			record := &model.BrandAccountRecord{}
			record.Account = v.Name
			record.BeginTime = v.BTime
			record.EndTime = v.ETime
			record.Date = convert.ToInt64(function.GetDateUnix())
			record.AddRecord(*record)
		}
	}
}

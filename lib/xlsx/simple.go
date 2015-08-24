package xlsx

import "github.com/tealeg/xlsx"

//创建excel
func CreateXlsx(fileName string, headers []string, list [][]string) {
	f := xlsx.NewFile()
	sheet := f.AddSheet("sheet1")

	//头部
	rows := sheet.AddRow()
	for _, v := range headers {
		rows.AddCell().Value = v
	}

	//数据
	for _, v := range list {
		rows := sheet.AddRow()
		for _, vv := range v {
			rows.AddCell().Value = vv
		}
	}

	//保存
	f.Save(fileName)
}

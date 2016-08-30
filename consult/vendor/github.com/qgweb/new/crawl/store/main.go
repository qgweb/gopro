package main

import "github.com/qgweb/new/crawl/store/model"

func main() {
	var ds model.DataStorer
	config := model.ParseConfig()
	switch config.GType {
	case "taobao":
		ds = model.NewTaoBaoDataStore(config)
	case "taobao_es":
		ds = model.NewTaoBaoEsDataStore(config)
	case "jd" :
		ds = model.NewJDDataStore(config)
	}

	ds.Receive(config.ReceiveKey, config.NsqHost, config.NsqPort)
}

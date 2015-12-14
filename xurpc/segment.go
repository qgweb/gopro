package main

import (
	"github.com/ngaut/log"
	"github.com/wangbin/jiebago/analyse"
)

type SearchWord struct{}

var seg analyse.TagExtracter

func (this SearchWord) GetKeyWord(word string) []byte {
	var err error
	err = seg.LoadDictionary("./dictionary/dict.txt")
	if err != nil {
		log.Fatal("打开字典文件错误")
	}
	err = seg.LoadIdf("./dictionary/idf.txt")
	if err != nil {
		log.Fatal("打开逆向字典文件错误")
	}

	result := seg.ExtractTags(word, 5)
	slice := []string{}
	if len(result) > 0 {
		for _, v := range result {
			log.Info(v.Text())
			slice = append(slice, v.Text())
		}
	}
	return jsonReturn(slice, nil)
}

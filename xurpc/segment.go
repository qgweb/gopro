package main

import (
	"github.com/ngaut/log"
)

type SearchWord struct{}

func (this SearchWord) GetKeyWord(word string) []byte {
	result := seg.ExtractTags(word, 5)
	slice := []string{}
	if len(result) > 0 {
		for _, v := range result {
			// log.Info(v.Text())
			slice = append(slice, v.Text())
		}
	}
	return jsonReturn(slice, nil)
}

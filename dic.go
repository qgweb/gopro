package main

import (
	"fmt"
	"log"

	"github.com/wangbin/jiebago/analyse"
)

var (
	seg analyse.TagExtracter
	err error
)

func init() {
	fmt.Println(123123)
	err = seg.LoadDictionary("/data/workgo/src/test/dict.txt")
	fmt.Println(err)
	if err != nil {
		fmt.Println(err)
		log.Fatal("打开字典文件错误")
	}
	err = seg.LoadIdf("./xurpc/dictionary/idf.txt")
	if err != nil {
		fmt.Println(err)
		log.Fatal("打开逆向字典文件错误")
	}
	fmt.Println("xxxx")
}

func print(ch <-chan string) {
	for word := range ch {
		fmt.Printf(" %s /", word)
	}
	fmt.Println()
}

func main() {
	return
}

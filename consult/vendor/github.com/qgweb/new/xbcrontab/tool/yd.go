package main

import (
	"bufio"
	"flag"
	"io"
	"log"
	"os"
	"strings"

	"github.com/qgweb/gopro/lib/encrypt"
	"github.com/qgweb/new/lib/rediscache"
)

var (
	dx_host    = flag.String("dxhost", "192.168.1.199", "")
	dx_port    = flag.String("dxport", "6380", "")
	dx_auth    = flag.String("dxauth", "", "")
	dx_db      = flag.String("dxdb", "14", "")
	put_host   = flag.String("phost", "192.168.1.199", "")
	put_port   = flag.String("pport", "6380", "")
	put_db     = flag.String("pdb", "3", "")
	tagHashKey = flag.String("tagkey", "", "")
	fname      = flag.String("file", "put.txt", "")
)

func init() {
	flag.Parse()
}

// 获取标签对应的广告集合
func getTagAdvertHash(tags string, taghash map[string]string) map[string]int {
	var ntags = strings.Split(tags, ",")
	var result = make(map[string]int)
	for _, v := range ntags {
		if aid, ok := taghash[v]; ok {
			result[aid] = 1
		}
	}
	return result
}

func main() {
	f, err := os.Open(*fname)
	if err != nil {
		log.Fatalln(err)
	}
	defer f.Close()

	put, err := rediscache.New(rediscache.MemConfig{Host: *put_host, Port: *put_port, Db: *put_db})
	if err != nil {
		log.Fatalln(err)
	}
	defer put.Close()

	dx, err := rediscache.New(rediscache.MemConfig{Host: *dx_host, Port: *dx_port, Db: *dx_db})
	if err != nil {
		log.Fatalln(err)
	}
	defer dx.Close()
	dx.Auth(*dx_auth)
	dx.SelectDb(*dx_db)

	tagAdverts := put.HGetAll(*tagHashKey)
	put.SelectDb(*put_db)

	bi := bufio.NewReader(f)
	for {
		line, err := bi.ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			break
		}

		datas := strings.Split(strings.TrimSpace(line), "\t")
		log.Println(len(datas))
		if len(datas) < 3 {
			continue
		}

		ad := datas[0]
		ua := datas[1]
		tags := getTagAdvertHash(datas[2], tagAdverts)
		log.Println(tags)
		put_key := encrypt.DefaultMd5.Encode(ad + "_" + ua)
		log.Println(put_key)
		dx_key := ad + "|" + strings.ToUpper(ua)
		log.Println(dx_key)
		for aid, _ := range tags {
			put.HSet(put_key, "advert:"+aid, aid)
			dx.Set(dx_key, "1")
		}
		put.Expire(put_key, 5400)

	}
	put.Flush()
	put.Flush()
	dx.Flush()
}

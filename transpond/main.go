//接收信号量,转发到其他程序
package main

import (
	"bufio"
	"flag"
	"io"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Unknwon/goconfig"
	"github.com/garyburd/redigo/redis"
)

var (
	conf  = flag.String("conf", "conf.ini", "配置文件")
	fdate = flag.String("date", "", "读取文件时间")
	ini   *goconfig.ConfigFile
	rhost string
	rport string
	err   error
	pool  *redis.Pool
)

func init() {
	flag.Parse()

	if *conf == "" {
		log.Fatalln("配置文件不能为空")
	}

	ini, err = goconfig.LoadConfigFile(*conf)
	if err != nil {
		log.Fatalln("读取配置文件出错,错误信息为:", err)
	}

	rhost, err = ini.GetValue("redis", "host")
	if err != nil || rhost == "" {
		log.Fatalln("读取redis-host出错")
	}

	rport, err = ini.GetValue("redis", "port")
	if err != nil || rhost == "" {
		log.Fatalln("读取redis-port出错")
	}

	pool = RedisPool()
}

//读取文件
func readFile() {
	filePath, err := ini.GetValue("pro", "datapath")
	if err != nil || filePath == "" {
		log.Fatalln("文件目录不存在")
	}

	if *fdate == "" {
		*fdate = time.Now().Add(-time.Second * 86400).Format("20060102")
	}

	fileName := filePath + "/" + *fdate
	f, err := os.Open(fileName)
	if err != nil {
		log.Fatalln("目标文件不存在")
	}

	hurl, err := ini.GetValue("pro", "httpsqs")
	if err != nil || hurl == "" {
		log.Fatalln("转发地址不存在")
	}

	bi := bufio.NewReaderSize(f, 1<<26)

	for {
		line, err := bi.ReadString('\n')
		if err == io.EOF || line == "\n" {
			break
		}

		//分发数据
		dispath(strings.TrimSpace(line), hurl)
	}

}

//获取reids连接池
func RedisPool() *redis.Pool {
	return &redis.Pool{
		MaxIdle:   100,
		MaxActive: 100, // max number of connections
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", rhost+":"+rport)
			if err != nil {
				panic(err.Error())
			}
			return c, err
		},
	}
}

//发送数据
func sendData(param url.Values) {
	conn := pool.Get()
	defer conn.Close()

	prefix, _ := ini.GetValue("pro", "prefix")

	_, err := conn.Do("LPUSH", prefix+"taglist", param.Encode())
	if err != nil {
		log.Println("写入redis出错")
	}
}

//分配数据
func dispath(data string, uri string) {
	datas := strings.Split(data, "\t")

	dataIndex, err := ini.GetSection("data")
	if err != nil {
		log.Fatalln("数据顺序小节没有")
	}

	u := url.Values{}

	for k, v := range dataIndex {
		vi, _ := strconv.Atoi(v)
		if len(datas) > vi {
			u.Add(k, datas[vi])
		}
	}

	sendData(u)
}

func main() {
	//读取pid文件
	pid, err := ini.GetValue("pro", "pid")
	if err != nil || pid == "" {
		pid = "./transpond.pid"
	}

	ioutil.WriteFile(pid, []byte(strconv.Itoa(os.Getpid())), 0755)

	readFile()
}

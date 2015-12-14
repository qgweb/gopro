package redis

import (
	"log"

	"github.com/garyburd/redigo/redis"
	"github.com/qgweb/gopro/qianzhao/common/config"
)

var (
	pool *redis.Pool
)

func init() {
	var (
		host = config.GetRedis().Key("host").String()
		port = config.GetRedis().Key("port").String()
	)

	pool = &redis.Pool{
		MaxIdle:   10,
		MaxActive: 10, // max number of connections
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", host+":"+port)
			if err != nil {
				log.Fatalln(err.Error())
			}
			return c, err
		},
	}
}

// 获取连接
func Get() redis.Conn {
	return pool.Get()
}
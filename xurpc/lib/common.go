package lib

import (
	"encoding/json"
	"github.com/garyburd/redigo/redis"
)

func JsonReturn(data interface{}, err error) []byte {
	type Ret struct {
		Ret  int
		Msg  string
		Data interface{}
	}
	if err != nil {
		d, _ := json.Marshal(&Ret{Ret: 1, Msg: err.Error(), Data: nil})
		return d
	}

	d, _ := json.Marshal(&Ret{Ret: 0, Msg: "", Data: data})
	return d
}

func GetRedisPool(host, port string) *redis.Pool {
	return &redis.Pool{
		MaxIdle:   80,
		MaxActive: 1200, // max number of connections
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", host+":"+port)
			if err != nil {
				panic(err.Error())
			}
			return c, err
		},
	}
}

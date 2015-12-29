package xredis

import (
	"fmt"
	"github.com/astaxie/beego"
	"github.com/garyburd/redigo/redis"
	"os"
	"time"
	"sync"
)

var (
	pool    *redis.Pool
	rconfig *RedisConfig
	mux sync.RWMutex
)

type RedisConfig struct {
	Host string
	Port string
	Auth string
	Db string
}

func init() {
	rconfig = getRedisConfig()
	pool = newPool(fmt.Sprintf("%s:%s", rconfig.Host, rconfig.Port), rconfig.Auth)
}

func getRedisConfig() *RedisConfig {
	mc := &RedisConfig{}
	config, err := beego.AppConfig.GetSection("redis")
	if err != nil {
		beego.Emergency(err)
		os.Exit(-2)
	}
	if v, ok := config["host"]; ok {
		mc.Host = v
	}
	if v, ok := config["port"]; ok {
		mc.Port = v
	}
	if v, ok := config["auth"]; ok {
		mc.Auth = v
	}
	if v, ok := config["db"]; ok {
		mc.Db = v
	}
	return mc
}

func newPool(server, password string) *redis.Pool {
	return &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", server)
			if err != nil {
				return nil, err
			}
			if password != "" {
				if _, err := c.Do("AUTH", password); err != nil {
					c.Close()
					return nil, err
				}
			}
			return c, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
}

func GetRedis() redis.Conn {
	mux.Lock()
	defer mux.Unlock()
	return pool.Get()
}

func GetConfig() *RedisConfig {
	mux.RLock()
	defer mux.RUnlock()
	return rconfig
}
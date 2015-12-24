package rediscache

import (
	"fmt"
	"github.com/garyburd/redigo/redis"
)

type MemConfig struct {
	Host string
	Port string
}

type MemCache struct {
	db redis.Conn
}

func New(config MemConfig) (*MemCache, error) {
	conn, err := redis.Dial("tcp4", fmt.Sprintf("%s:%s", config.Host, config.Port))
	if err != nil {
		return nil, err
	}
	return &MemCache{conn}, nil
}

func (this *MemCache) Get(key string) string {
	if res, err := redis.String(this.db.Do("GET", key)); err == nil {
		return res
	}
	return ""
}

func (this *MemCache) Expire(key string, seconds int) {
	this.db.Send("EXPIRE", key, seconds)
}

func (this *MemCache) Set(key string, val string) {
	this.db.Send("SET", key, val)
}

func (this *MemCache) Del(key string) {
	this.db.Send("DEL", key)
}

func (this *MemCache) HGet(key string, name string) string {
	if res, err := redis.String(this.db.Do("HGET", key, name)); err == nil {
		return res
	}
	return ""
}

func (this *MemCache) HSet(key string, name string, val string) {
	this.db.Send("HSET", key, name, val)
}

func (this *MemCache) Has(key string) bool {
	if res, err := redis.Bool(this.db.Do("EXISTS", key)); err == nil {
		return res
	}
	return false
}

func (this *MemCache) HDel(key string, name string) {
	this.db.Send("HDEL", key, name)
}

func (this *MemCache) HGetAllKeys(key string) []string {
	if res, err := redis.Strings(this.db.Do("HKEYS", key)); err == nil {
		return res
	}
	return make([]string, 0)
}

func (this *MemCache) HGetAllValue(key string) []string {
	if res, err := redis.Strings(this.db.Do("HVALS", key)); err == nil {
		return res
	}
	return make([]string, 0)
}

func (this *MemCache) Keys(key string) []string {
	if res, err := redis.Strings(this.db.Do("KEYS", key)); err == nil {
		return res
	}
	return make([]string, 0)
}

func (this *MemCache) SelectDb(db string) {
	this.db.Do("SELECT", db)
}

func (this *MemCache) Flush() {
	this.db.Flush()
}

func (this *MemCache) Close() {
	this.db.Close()
}

func (this *MemCache) Clean(prefix string) {
	keys := this.Keys(prefix + "*")
	for _, key := range keys {
		this.Del(key)
	}
	if len(keys) > 0 {
		this.db.Flush()
	}
}

package rediscache

import (
	"fmt"

	"github.com/garyburd/redigo/redis"
	"github.com/juju/errors"
)

type MemConfig struct {
	Host string
	Port string
	Db   string
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

func NewMul(config MemConfig, size int) ([]*MemCache, error) {
	if size <= 0 {
		return nil, errors.New("size不能小于0")
	}
	ms := make([]*MemCache, 0, size)
	for i := 0; i < size; i++ {
		if db, err := New(config); err != nil {
			return nil, err
		} else {
			ms = append(ms, db)
		}
	}
	return ms, nil
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

func (this *MemCache) Auth(pwd string) {
	this.db.Do("Auth", pwd)
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

func (this *MemCache) HGetAll(key string) map[string]string {
	if res, err := redis.StringMap(this.db.Do("HGETALL", key)); err == nil {
		return res
	}
	return make(map[string]string)
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

func (this *MemCache) FlushDb() {
	this.db.Do("FLUSHDB")
}

func (this *MemCache) SMembers(key string) []string {
	if res, err := redis.Strings(this.db.Do("SMEMBERS", key)); err == nil {
		return res
	}
	return make([]string, 0)
}

func (this *MemCache) Sadd(key string, values interface{}) (int, error) {
	return redis.Int(this.db.Do("SADD", key, values))
}

func (this *MemCache) Srem(key string, values interface{}) (int, error) {
	return redis.Int(this.db.Do("SREM", key, values))
}

func (this *MemCache) Rpush(key string, value string) (int, error) {
	return redis.Int(this.db.Do("RPUSH", key, value))
}

func (this *MemCache) Lpop(key string) (interface{}, error) {
	return redis.String(this.db.Do("LPOP", key))
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

func (this *MemCache) Receive() (interface{}, error) {
	return this.db.Receive()
}
// LevelDb 缓存类
// @author: zhengbo
// 性能: wirte 125000/s , read 500000/s
// 注意: keys 方法不同redis的keys, hset 会把name和key合成一个主key
//      想知道到hset所有的key, 可以再用hset存插入的key

package cache

import (
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/filter"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/syndtr/goleveldb/leveldb/util"
	"os"
	"strings"
)

type LevelDBCache struct {
	db     *leveldb.DB
	dbpath string
}

func NewLevelDbCache(path string) (*LevelDBCache, error) {
	o := &opt.Options{
		Filter:             filter.NewBloomFilter(10),
		BlockCacheCapacity: 64,
		WriteBuffer:        256 * 1024 * 1024,
	}
	db, err := leveldb.OpenFile(path, o)
	if err != nil {
		return nil, err
	}
	return &LevelDBCache{db, path}, nil

}

func (this *LevelDBCache) Set(key string, value string) error {
	return this.db.Put([]byte(key), []byte(value), nil)
}

func (this *LevelDBCache) Get(key string) (string, error) {
	v, err := this.db.Get([]byte(key), nil)
	return string(v), err
}

func (this *LevelDBCache) Del(key string) error {
	return this.db.Delete([]byte(key), nil)
}

func (this *LevelDBCache) Has(key string) bool {
	if v, err := this.db.Has([]byte(key), nil); err == nil {
		return v
	}
	return false
}

func (this *LevelDBCache) HSet(hname string, hkey string, hvalue string) error {
	return this.db.Put([]byte(hname+"_"+hkey), []byte(hvalue), nil)
}

func (this *LevelDBCache) HGet(hname string, hkey string) (string, error) {
	v, err := this.db.Get([]byte(hname+"_"+hkey), nil)
	return string(v), err
}

func (this *LevelDBCache) HDel(hname string) error {
	batch := new(leveldb.Batch)
	if keys, err := this.Keys(hname + "_"); err == nil {
		for _, v := range keys {
			batch.Delete([]byte(v))
		}
	}
	return this.db.Write(batch, nil)
}

func (this *LevelDBCache) HGetAllValue(hname string) ([]string, error) {
	var hvalues = make([]string, 0, 100)
	iter := this.db.NewIterator(util.BytesPrefix([]byte(hname+"_")), nil)
	for iter.Next() {
		hvalues = append(hvalues, string(iter.Value()))
	}
	iter.Release()
	err := iter.Error()
	return hvalues, err
}

func (this *LevelDBCache) HGetAllKeys(hname string) ([]string, error) {
	var hvalues = make([]string, 0, 100)
	iter := this.db.NewIterator(util.BytesPrefix([]byte(hname+"_")), nil)
	for iter.Next() {
		hvalues = append(hvalues, strings.TrimPrefix(string(iter.Key()), hname+"_"))
	}
	iter.Release()
	err := iter.Error()
	return hvalues, err
}

func (this *LevelDBCache) Keys(prefix string) ([]string, error) {
	var hvalues = make([]string, 0, 100)
	iter := this.db.NewIterator(util.BytesPrefix([]byte(prefix)), nil)
	for iter.Next() {
		hvalues = append(hvalues, string(iter.Key()))
	}
	iter.Release()
	err := iter.Error()
	return hvalues, err
}

func (this *LevelDBCache) Close() {
	this.db.Close()
	os.RemoveAll(this.dbpath)
}

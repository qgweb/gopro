package common

import (
	"github.com/ngaut/log"
	"github.com/seefan/gossdb"
)

type SSDBConfig struct {
	Host string `json:"host"`
	Port int    `json:"port"`
	Auth string `json:"auth"`
}

type SSDBPool struct {
	pool *gossdb.Connectors
}

func NewSsdbPool(conf *SSDBConfig) *SSDBPool {
	pool, err := gossdb.NewPool(&gossdb.Config{
		Host:             conf.Host,
		Port:             conf.Port,
		MinPoolSize:      5,
		MaxPoolSize:      50,
		AcquireIncrement: 5,
	})
	if err != nil {
		log.Error(err)
		return nil
	}

	return &SSDBPool{pool}
}

func (this *SSDBPool) Get() (*gossdb.Client, error) {
	return this.pool.NewClient()
}

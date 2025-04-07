package cache

import (
	"github.com/dgraph-io/ristretto/v2"
)

type ZonesCache struct {
	cacheInst *ristretto.Cache[string, []string]
}

func NewZonesCache() (*ZonesCache, error) {
	cacheInst, err := ristretto.NewCache(&ristretto.Config[string, []string]{
		NumCounters: 1 << 20,
		MaxCost:     1 << 20,
		BufferItems: 64,
	})
	if err != nil {
		return nil, err
	}
	ret := &ZonesCache{
		cacheInst: cacheInst,
	}
	return ret, nil
}

func (c *ZonesCache) Get(key string) ([]string, bool) {
	return c.cacheInst.Get(key)
}

func (c *ZonesCache) Set(key string, value []string) {
	c.cacheInst.Set(key, value, int64(len(value)))
}

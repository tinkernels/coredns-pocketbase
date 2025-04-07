package cache

import (
	"time"

	"github.com/dgraph-io/ristretto/v2"
	m "github.com/tinkernels/coredns-pocketbase/handler/pocketbase/model"
)

type RecordsCache struct {
	cacheInst *ristretto.Cache[string, []*m.Record]
}

func NewRecordsCache(capacity int) (*RecordsCache, error) {
	cacheInst, err := ristretto.NewCache(&ristretto.Config[string, []*m.Record]{
		NumCounters: int64(capacity),
		MaxCost:     int64(capacity),
		BufferItems: 64,
	})
	if err != nil {
		return nil, err
	}
	return &RecordsCache{
		cacheInst: cacheInst,
	}, nil
}

func (c *RecordsCache) Get(key string) ([]*m.Record, bool) {
	return c.cacheInst.Get(key)
}

func (c *RecordsCache) Set(key string, value []*m.Record) {
	minttl := uint32(0)
	for _, rec := range value {
		if rec.Ttl < minttl || minttl == 0 {
			minttl = rec.Ttl
		}
	}
	c.cacheInst.SetWithTTL(key, value, int64(len(value)), time.Duration(minttl)*time.Second)
}

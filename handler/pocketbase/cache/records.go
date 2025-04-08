// Package cache provides caching functionality for DNS records and zones.
package cache

import (
	"time"

	"github.com/dgraph-io/ristretto/v2"
	m "github.com/tinkernels/coredns-pocketbase/handler/pocketbase/model"
)

// RecordsCache provides caching for DNS records using Ristretto cache.
// It stores a mapping of cache keys to lists of DNS records.
type RecordsCache struct {
	cacheInst *ristretto.Cache[string, []*m.Record]
}

// NewRecordsCache creates a new RecordsCache instance with the specified capacity.
// The capacity determines the maximum number of items that can be stored in the cache.
// Returns the cache instance and any error encountered during initialization.
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

// Get retrieves a list of records from the cache for the given key.
// Returns the records and a boolean indicating if the key was found.
func (c *RecordsCache) Get(key string) ([]*m.Record, bool) {
	return c.cacheInst.Get(key)
}

// Set stores a list of records in the cache with the given key.
// The TTL is determined by the minimum TTL among all records in the list.
// The cost is calculated based on the length of the records slice.
func (c *RecordsCache) Set(key string, value []*m.Record) {
	minttl := uint32(0)
	for _, rec := range value {
		if rec.Ttl < minttl || minttl == 0 {
			minttl = rec.Ttl
		}
	}
	c.cacheInst.SetWithTTL(key, value, int64(len(value)), time.Duration(minttl)*time.Second)
}

func (c *RecordsCache) Delete(key string) {
	c.cacheInst.Del(key)
}

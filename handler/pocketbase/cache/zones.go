// Package cache provides caching functionality for DNS records and zones.
package cache

import (
	"github.com/dgraph-io/ristretto/v2"
)

// ZonesCache provides caching for DNS zones using Ristretto cache.
// It stores a mapping of cache keys to lists of zone names.
type ZonesCache struct {
	cacheInst *ristretto.Cache[string, []string]
}

// NewZonesCache creates a new ZonesCache instance with default configuration.
// Returns the cache instance and any error encountered during initialization.
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

// Get retrieves a list of zones from the cache for the given key.
// Returns the zones and a boolean indicating if the key was found.
func (c *ZonesCache) Get(key string) ([]string, bool) {
	return c.cacheInst.Get(key)
}

// Set stores a list of zones in the cache with the given key.
// The cost is calculated based on the length of the zones slice.
func (c *ZonesCache) Set(key string, value []string) {
	c.cacheInst.Set(key, value, int64(len(value)))
}

func (c *ZonesCache) Delete(key string) {
	c.cacheInst.Del(key)
}

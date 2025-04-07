package cache

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	m "github.com/tinkernels/coredns-pocketbase/handler/pocketbase/model"
)

func TestZonesCache(t *testing.T) {
	// Create a new cache
	cache, err := NewZonesCache()
	assert.NoError(t, err)

	// Test data
	testKey := "test_key"
	testZones := []string{"example.com.", "test.com."}

	// Test Set and Get
	cache.Set(testKey, testZones)
	time.Sleep(time.Millisecond * 1)
	zones, ok := cache.Get(testKey)

	assert.True(t, ok)
	assert.Equal(t, testZones, zones)

	// Test non-existent key
	_, ok = cache.Get("non_existent")
	assert.False(t, ok)
}

func TestRecordsCache(t *testing.T) {
	// Create a new cache with capacity of 1000
	cache, err := NewRecordsCache(1000)
	assert.NoError(t, err)

	// Test data
	testKey := "test_key"
	testRecords := []*m.Record{
		{
			Zone:       "example.com.",
			Name:       "www",
			RecordType: "A",
			Ttl:        300,
			Content:    `{"ip":"1.1.1.1"}`,
		},
	}

	// Test Set and Get
	cache.Set(testKey, testRecords)
	time.Sleep(time.Millisecond * 1)
	records, ok := cache.Get(testKey)
	assert.True(t, ok)
	assert.Equal(t, testRecords, records)

	// Test non-existent key
	_, ok = cache.Get("non_existent")
	assert.False(t, ok)
}

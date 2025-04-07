// Package handler provides configuration management for the CoreDNS PocketBase integration.
// It defines the configuration structure and default values for the service.
package handler

import (
	"fmt"
	"net"
	"os"
)

// Default configuration values
const (
	// defaultListen is the default address and port the service listens on
	defaultListen = "[::]:8090"
	// defaultDataDir is the default directory for PocketBase data storage
	defaultDataDir = "pb_data"
	// defaultSuEmail is the default superuser email for PocketBase
	defaultSuEmail = "su@pocketbase.internal"
	// defaultSuPassword is the default superuser password for PocketBase
	defaultSuPassword = "pwd@pocketbase.internal"
	// defaultCacheCapacity is the default number of records to cache (0 means no caching)
	defaultCacheCapacity = 0
	// DefaultDefaultTtl is the default TTL (Time To Live) in seconds for DNS records
	defaultDefaultTtl = 30
)

// Config represents the configuration for the CoreDNS PocketBase integration.
// It contains settings for the service's network interface, data storage,
// authentication, and caching behavior.
type Config struct {
	// Listen is the address and port the service listens on
	Listen string
	// DataDir is the directory for PocketBase data storage
	DataDir string
	// SuEmail is the superuser email for PocketBase
	SuEmail string
	// SuPassword is the superuser password for PocketBase
	SuPassword string
	// CacheCapacity is the number of records to cache (0 means no caching)
	CacheCapacity int
	// DefaultTtl is the default TTL (Time To Live) in seconds for DNS records
	DefaultTtl int
}

// NewConfig creates a new Config instance with default values
func NewConfig() *Config {
	return &Config{
		Listen:        defaultListen,
		DataDir:       defaultDataDir,
		SuEmail:       defaultSuEmail,
		SuPassword:    defaultSuPassword,
		CacheCapacity: defaultCacheCapacity,
		DefaultTtl:    defaultDefaultTtl,
	}
}

func DefaultConfigVal4DefaultTtl() int {
	return defaultDefaultTtl
}

// WithListen sets the listen address and returns the modified Config
func (c *Config) WithListen(listen string) *Config {
	c.Listen = listen
	return c
}

// WithDataDir sets the data directory and returns the modified Config
func (c *Config) WithDataDir(dataDir string) *Config {
	c.DataDir = dataDir
	return c
}

// WithSuEmail sets the superuser email and returns the modified Config
func (c *Config) WithSuEmail(suUserName string) *Config {
	c.SuEmail = suUserName
	return c
}

// WithSuPassword sets the superuser password and returns the modified Config
func (c *Config) WithSuPassword(suPassword string) *Config {
	c.SuPassword = suPassword
	return c
}

// WithCacheCapacity sets the cache capacity and returns the modified Config
func (c *Config) WithCacheCapacity(cacheCapacity int) *Config {
	c.CacheCapacity = cacheCapacity
	return c
}

// WithDefaultTtl sets the default TTL and returns the modified Config
func (c *Config) WithDefaultTtl(defaultTtl int) *Config {
	c.DefaultTtl = defaultTtl
	return c
}

func (c *Config) MixWithEnv() *Config {
	if suUserName := os.Getenv("COREDNS_PB_SUPERUSER_EMAIL"); suUserName != "" {
		c.SuEmail = suUserName
	}
	if suPassword := os.Getenv("COREDNS_PB_SUPERUSER_PWD"); suPassword != "" {
		c.SuPassword = suPassword
	}
	return c
}

// Validate checks if the configuration is valid
// Returns an error if any required field is empty or if numeric values are negative
func (c *Config) Validate() error {
	if c.Listen == "" {
		return fmt.Errorf("listen is required")
	}
	if _, err := net.ResolveTCPAddr("tcp", c.Listen); err != nil {
		return fmt.Errorf("invalid listen address format: %v", err)
	}
	if c.DataDir == "" {
		return fmt.Errorf("data_dir is required")
	}
	if c.SuEmail == "" {
		return fmt.Errorf("su_email is required")
	}
	if c.SuPassword == "" {
		return fmt.Errorf("su_password is required")
	}
	if c.CacheCapacity < 0 {
		return fmt.Errorf("cache_capacity must be greater than or equal to 0")
	}
	if c.DefaultTtl < 0 {
		return fmt.Errorf("default_ttl must be greater than or equal to 0")
	}
	return nil
}

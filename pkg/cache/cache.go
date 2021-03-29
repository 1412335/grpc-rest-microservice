package cache

import (
	"context"
	"time"
)

// cacheDefaultExpiration defaults time for a value in the cache to expire
const cacheDefaultExpiration = 1 * time.Hour

// cacheMaxListLimit defines maximum number of items to pull from the cache when doing a list of keys
const cacheLRUMaxSize = 200

type Options func(*Option) error

func WithDatabase(database string) Options {
	return func(c *Option) error {
		c.database = database
		return nil
	}
}

func WithPrefix(prefix string) Options {
	return func(c *Option) error {
		c.prefix = prefix
		return nil
	}
}

func WithExpiryDuration(expiryDuration time.Duration) Options {
	return func(c *Option) error {
		if expiryDuration == 0 {
			expiryDuration = cacheDefaultExpiration
		}
		c.expiryDuration = expiryDuration
		return nil
	}
}

func WithLRUMaxSize(size int) Options {
	return func(c *Option) error {
		if size == 0 {
			size = cacheLRUMaxSize
		}
		c.lruMaxSize = size
		return nil
	}
}

type Option struct {
	ctx            context.Context
	database       string
	prefix         string
	expiryDuration time.Duration
	lruMaxSize     int
}

type Cache interface {
	Close() error
	Set(key, value string) error
	Get(key string, val interface{}) error
	Delete(key string) error
}

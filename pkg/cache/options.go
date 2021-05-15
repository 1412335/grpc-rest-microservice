package cache

import (
	"context"
	"time"
)

// cacheDefaultExpiration defaults time for a value in the cache to expire
const cacheDefaultExpiration = 1 * time.Hour

// cacheMaxListLimit defines maximum number of items to pull from the cache when doing a list of keys
const cacheLRUMaxSize = 200

type Option func(*Options) error

func WithDatabase(database string) Option {
	return func(c *Options) error {
		c.database = database
		return nil
	}
}

func WithPrefix(prefix string) Option {
	return func(c *Options) error {
		c.prefix = prefix
		return nil
	}
}

func WithExpiryDuration(expiryDuration time.Duration) Option {
	return func(c *Options) error {
		if expiryDuration == 0 {
			expiryDuration = cacheDefaultExpiration
		}
		c.expiryDuration = expiryDuration
		return nil
	}
}

func WithLRUMaxSize(size int) Option {
	return func(c *Options) error {
		if size == 0 {
			size = cacheLRUMaxSize
		}
		c.lruMaxSize = size
		return nil
	}
}

type Options struct {
	ctx            context.Context
	database       string
	prefix         string
	expiryDuration time.Duration
	lruMaxSize     int
}

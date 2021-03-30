package cache

import (
	"context"
	"time"

	"github.com/1412335/grpc-rest-microservice/pkg/dal/redis"
	rdCache "github.com/go-redis/cache/v8"
)

type redisCache struct {
	opts  Option
	cache *rdCache.Cache
}

var _ Cache = (*redisCache)(nil)

func NewRedisCache(store *redis.Redis, opts ...Options) (Cache, error) {
	c := redisCache{
		opts: Option{
			ctx:            context.Background(),
			database:       "",
			prefix:         "",
			expiryDuration: cacheDefaultExpiration,
			lruMaxSize:     cacheLRUMaxSize,
		},
	}
	for _, o := range opts {
		if err := o(&c.opts); err != nil {
			return nil, err
		}
	}
	cache := rdCache.New(&rdCache.Options{
		Redis:      store.GetClient(),
		LocalCache: rdCache.NewTinyLFU(c.opts.lruMaxSize, 1*time.Minute),
	})
	c.cache = cache
	return &c, nil
}

func (c *redisCache) Close() error {
	return nil
}

func (c *redisCache) Set(key, value string) error {
	err := c.cache.Set(&rdCache.Item{
		Ctx:   c.opts.ctx,
		Key:   key,
		Value: []byte(value),
		TTL:   c.opts.expiryDuration,
	})
	return err
}

func (c *redisCache) Get(key string, obj interface{}) error {
	err := c.cache.Get(c.opts.ctx, key, obj)
	return err
}

func (c *redisCache) Delete(key string) error {
	err := c.cache.Delete(c.opts.ctx, key)
	return err
}

func (c *redisCache) Exists(key string) bool {
	return c.cache.Exists(c.opts.ctx, key)
}

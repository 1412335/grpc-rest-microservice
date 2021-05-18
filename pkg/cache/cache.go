package cache

import "errors"

var (
	DefaultCache         Cache
	ErrCacheNotAvailable = errors.New("cache not available")
)

type Cache interface {
	Close() error
	Set(key, value string) error
	Get(key string, val interface{}) error
	Delete(key string) error
	Ratio() float64
}

func Set(key, value string) error {
	if DefaultCache == nil {
		return ErrCacheNotAvailable
	}
	return DefaultCache.Set(key, value)
}

func Get(key string, val interface{}) error {
	if DefaultCache == nil {
		return ErrCacheNotAvailable
	}
	return DefaultCache.Get(key, val)
}

func Delete(key string) error {
	if DefaultCache == nil {
		return ErrCacheNotAvailable
	}
	return DefaultCache.Delete(key)
}

func Close() error {
	if DefaultCache == nil {
		return ErrCacheNotAvailable
	}
	return DefaultCache.Close()
}

func Ratio() float64 {
	if DefaultCache == nil {
		return 0.0
	}
	return DefaultCache.Ratio()
}

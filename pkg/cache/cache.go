package cache

var (
	DefaultCache Cache
)

type Cache interface {
	Close() error
	Set(key, value string) error
	Get(key string, val interface{}) error
	Delete(key string) error
	Ratio() float64
}

func Set(key, value string) error {
	return DefaultCache.Set(key, value)
}

func Get(key string, val interface{}) error {
	return DefaultCache.Get(key, val)
}

func Delete(key string) error {
	return DefaultCache.Delete(key)
}

func Close() error {
	return DefaultCache.Close()
}

func Ratio() float64 {
	return DefaultCache.Ratio()
}

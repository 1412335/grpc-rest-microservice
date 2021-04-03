package redis

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	redis "github.com/go-redis/redis/v8"
)

//ctx context to be used when interacting with Redis
var ctx = context.Background()
var once sync.Once

// rkv is the struct that handles the client connected to Redis
type Redis struct {
	nodes  []string
	prefix string
	client *redis.Client
}

//NewStore creates and returns a rkv redis object
func New(opts ...Option) (*Redis, error) {
	r := &Redis{}
	for _, opt := range opts {
		if err := opt(r); err != nil {
			return nil, err
		}
	}
	if err := r.Connect(); err != nil {
		return nil, err
	}
	return r, nil
}

//configure set the address of the redis instance
func (r *Redis) configure() error {
	var redisOptions *redis.Options
	nodes := r.nodes
	if len(nodes) == 0 {
		nodes = []string{"redis://127.0.0.1:6379"}
	}

	redisOptions, err := redis.ParseURL(nodes[0])
	if err != nil {
		//Backwards compatibility
		redisOptions = &redis.Options{
			Addr:        nodes[0],
			Password:    "",          // no password set
			DB:          0,           // use default DB
			PoolTimeout: time.Minute, //
		}
	}

	r.client = redis.NewClient(redisOptions)
	return nil
}

// Connect
func (r *Redis) Connect() error {
	//Perform connection creation operation only once.
	var e error
	once.Do(func() {
		if err := r.configure(); err != nil {
			e = err
			return
		}
	})
	return e
}

// get redis client
func (r *Redis) GetClient() *redis.Client {
	if err := r.Connect(); err != nil {
		return nil
	}
	return r.client
}

//Close ends the connection to Redis
func (r *Redis) Close() error {
	return r.client.Close()
}

//Read retrieves data from Redis based on a given key
func (r *Redis) Read(key string, opts ...ReadOption) ([]*Record, error) {
	var rOpts ReadOptions
	rOpts.Prefix = r.prefix
	for _, opt := range opts {
		if err := opt(&rOpts); err != nil {
			return nil, err
		}
	}

	var keys []string
	rkey := fmt.Sprintf("%s%s", rOpts.Prefix, key)
	// Handle Prefix
	// TODO suffix
	if r.prefix != "" {
		prefixKey := fmt.Sprintf("%s*", rkey)
		fkeys, err := r.client.Keys(ctx, prefixKey).Result()
		if err != nil {
			return nil, err
		}
		// TODO Limit Offset
		keys = append(keys, fkeys...)
	} else {
		keys = []string{rkey}
	}

	records := make([]*Record, 0, len(keys))
	for _, rkey = range keys {
		val, err := r.client.Get(ctx, rkey).Bytes()

		if err != nil && err == redis.Nil {
			return nil, errors.New("not found")
		} else if err != nil {
			return nil, err
		}
		if val == nil {
			return nil, errors.New("not found")
		}
		d, err := r.client.TTL(ctx, rkey).Result()
		if err != nil {
			return nil, err
		}
		records = append(records, &Record{
			Key:    key,
			Value:  val,
			Expiry: d,
		})
	}
	return records, nil
}

//Delete removes data from Redis based on the given Key
func (r *Redis) Del(key string, opts ...DeleteOption) error {
	var dOpts DeleteOptions
	dOpts.Prefix = r.prefix
	for _, opt := range opts {
		if err := opt(&dOpts); err != nil {
			return err
		}
	}
	rkey := fmt.Sprintf("%s%s", dOpts.Prefix, key)
	return r.client.Del(ctx, rkey).Err()
}

//Write save data to redis
func (r *Redis) Write(record *Record, opts ...WriteOption) error {
	var wOpts WriteOptions
	wOpts.Prefix = r.prefix
	wOpts.Expiry = record.Expiry
	for _, opt := range opts {
		if err := opt(&wOpts); err != nil {
			return err
		}
	}
	rkey := fmt.Sprintf("%s%s", wOpts.Prefix, record.Key)
	return r.client.Set(ctx, rkey, record.Value, wOpts.Expiry).Err()
}

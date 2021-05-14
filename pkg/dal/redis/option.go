package redis

import (
	"time"
)

type Option func(*Redis) error

func WithNodes(nodes []string) Option {
	return func(r *Redis) error {
		r.nodes = nodes
		return nil
	}
}

func WithPrefix(prefix string) Option {
	return func(r *Redis) error {
		r.prefix = prefix
		return nil
	}
}

type ReadOption func(*ReadOptions) error

type ReadOptions struct {
	Prefix string
	Limit  int
	Offset int
}

type DeleteOption func(*DeleteOptions) error

type DeleteOptions struct {
	Prefix string
}

type WriteOption func(*WriteOptions) error

type WriteOptions struct {
	Prefix string
	Expiry time.Duration
}

type Record struct {
	Key    string        `json:"key"`
	Value  interface{}   `json:"value"`
	Expiry time.Duration `json:"expiry"`
}

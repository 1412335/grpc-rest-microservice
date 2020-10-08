package bridge

import (
	"context"
	api_v2 "grpc-rest-microservice/pkg/api/v2/gen/grpc-gateway/gen"

	grpcpool "github.com/processout/grpc-go-pool"
)

type Client interface {
	Ping(timestamp int64) (*api_v2.MessagePong, error)
	Post(timestamp int64) (*api_v2.MessagePong, error)

	// StreamingPing(timestamp int64, count, interval int32) error
}

type ClientImpl struct {
	client api_v2.ServiceExtraClient
	conn   *grpcpool.ClientConn
	ctx    context.Context
}

func (c *ClientImpl) Close() error {
	return c.conn.Close()
}

func (c *ClientImpl) Ping(timestamp int64) (*api_v2.MessagePong, error) {
	msg := &api_v2.MessagePing{
		Timestamp: timestamp,
	}
	return c.client.Ping(c.ctx, msg)
}

func (c *ClientImpl) Post(timestamp int64) (*api_v2.MessagePong, error) {
	msg := &api_v2.MessagePing{
		Timestamp: timestamp,
	}
	return c.client.Post(c.ctx, msg)
}

package bridge

import (
	"context"
	api_v2 "grpc-rest-microservice/pkg/api/v2/gen/grpc-gateway/gen"
	"io"
	"log"

	grpcpool "github.com/processout/grpc-go-pool"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type Client interface {
	Ping(timestamp int64) (*api_v2.MessagePong, error)
	Post(timestamp int64) (*api_v2.MessagePong, error)

	StreamingPing(timestamp int64, count, interval int32) ([]*api_v2.StreamingMessagePong, error)
}

type ClientImpl struct {
	client api_v2.ServiceExtraClient
	conn   *grpcpool.ClientConn
	ctx    context.Context
}

func (c *ClientImpl) Close() error {
	return c.conn.Close()
}

func (c *ClientImpl) setHeader(m map[string]string) (context.Context, error) {
	md := metadata.New(m)
	ctx := metadata.NewOutgoingContext(c.ctx, md)
	return ctx, nil
}

func (c *ClientImpl) Ping(timestamp int64) (*api_v2.MessagePong, error) {
	ctx, err := c.setHeader(map[string]string{"custom-req-header": "ping"})
	if err != nil {
		return nil, err
	}

	msg := &api_v2.MessagePing{
		Timestamp: timestamp,
	}

	// fetch response headers
	var header metadata.MD

	reply, err := c.client.Ping(ctx, msg, grpc.Header(&header))
	if err != nil {
		return nil, err
	}
	xrid := header.Get("x-response-id")
	if len(xrid) > 0 {
		log.Printf("'x-response-id': %v\n", xrid[0])
	}
	return reply, nil
}

func (c *ClientImpl) Post(timestamp int64) (*api_v2.MessagePong, error) {
	ctx, err := c.setHeader(map[string]string{"custom-req-header": "post"})
	if err != nil {
		return nil, err
	}

	msg := &api_v2.MessagePing{
		Timestamp: timestamp,
	}

	var header metadata.MD
	reply, err := c.client.Post(ctx, msg, grpc.Header(&header))
	if err != nil {
		return nil, err
	}
	xrid := header.Get("x-response-id")
	if len(xrid) > 0 {
		log.Printf("'x-response-id': %v\n", xrid[0])
	}
	return reply, nil
}

func (c *ClientImpl) StreamingPing(timestamp int64, count, interval int32) ([]*api_v2.StreamingMessagePong, error) {
	msg := &api_v2.StreamingMessagePing{
		Timestamp:       timestamp,
		MessageCount:    count,
		MessageInterval: interval,
	}
	stream, err := c.client.StreamingPing(c.ctx, msg)
	if err != nil {
		return nil, err
	}
	var i int32
	resp := make([]*api_v2.StreamingMessagePong, count)
	for {
		reply, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		resp[i] = reply
		i++
	}
	return resp, nil
}

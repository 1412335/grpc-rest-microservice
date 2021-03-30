package client

import (
	"context"
	"io"
	"log"
	"time"

	api_v2 "github.com/1412335/grpc-rest-microservice/pkg/api/v2/grpc-gateway/gen"

	grpcpool "github.com/processout/grpc-go-pool"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type Client interface {
	Login() (string, error)

	Ping(timestamp int64) (*api_v2.MessagePong, error)
	Post(timestamp int64) (*api_v2.MessagePong, error)

	StreamingPing(timestamp int64, count, interval int32) ([]*api_v2.StreamingMessagePong, error)
	StreamingPost(in []*api_v2.StreamingMessagePing) (*api_v2.StreamingMessagePong, error)
	DuplexStreaming(in []*api_v2.StreamingMessagePing) ([]*api_v2.StreamingMessagePong, error)
}

type clientImpl struct {
	client   api_v2.ServiceExtraClient
	conn     *grpcpool.ClientConn
	ctx      context.Context
	username string
	password string
}

func (c *clientImpl) Close() error {
	return c.conn.Close()
}

func (c *clientImpl) setHeader(m map[string]string) context.Context {
	md := metadata.New(m)
	ctx := metadata.NewOutgoingContext(c.ctx, md)
	return ctx
}

// login & get token
func (c *clientImpl) Login() (string, error) {
	ctx := c.setHeader(map[string]string{"custom-req-header": "login"})
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	msg := &api_v2.LoginRequest{
		Username: c.username,
		Password: c.password,
	}

	// fetch response headers
	var header metadata.MD

	reply, err := c.client.Login(ctx, msg, grpc.Header(&header))
	if err != nil {
		return "", err
	}
	token := header.Get("token")
	if len(token) > 0 {
		log.Printf("'token': %v\n", token[0])
	}

	return reply.GetToken(), nil
}

// unary get
func (c *clientImpl) Ping(timestamp int64) (*api_v2.MessagePong, error) {
	ctx := c.setHeader(map[string]string{"custom-req-header": "ping"})
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

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
	token := header.Get("token")
	if len(token) > 0 {
		log.Printf("'token': %v\n", token[0])
	}
	return reply, nil
}

// unary post
func (c *clientImpl) Post(timestamp int64) (*api_v2.MessagePong, error) {
	ctx := c.setHeader(map[string]string{"custom-req-header": "post"})

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

// server streaming
func (c *clientImpl) StreamingPing(timestamp int64, count, interval int32) ([]*api_v2.StreamingMessagePong, error) {
	ctx := c.setHeader(map[string]string{"custom-req-header": "server-streaming-ping"})

	msg := &api_v2.StreamingMessagePing{
		Timestamp:       timestamp,
		MessageCount:    count,
		MessageInterval: interval,
	}
	stream, err := c.client.StreamingPing(ctx, msg)
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

// client streaming
func (c *clientImpl) StreamingPost(in []*api_v2.StreamingMessagePing) (*api_v2.StreamingMessagePong, error) {
	ctx := c.setHeader(map[string]string{"custom-req-header": "client-streaming-post"})

	stream, err := c.client.StreamingPost(ctx)
	if err != nil {
		return nil, err
	}
	// send msg into stream
	for _, msg := range in {
		if err = stream.Send(msg); err != nil {
			return nil, stream.RecvMsg(nil)
		}
	}
	// close send stream
	if e := stream.CloseSend(); e != nil {
		return nil, e
	}
	// receive resp
	reply, err := stream.CloseAndRecv()
	if err != nil {
		return nil, err
	}
	return reply, nil
}

// bi-directional streaming
func (c *clientImpl) DuplexStreaming(in []*api_v2.StreamingMessagePing) ([]*api_v2.StreamingMessagePong, error) {
	ctx := c.setHeader(map[string]string{"custom-req-header": "duplex-streaming"})

	stream, err := c.client.DuplexStreamingPing(ctx)
	if err != nil {
		return nil, err
	}
	// send msg into stream
	errChan := make(chan error)
	go func(errChan chan<- error) {
		for _, msg := range in {
			if err := stream.Send(msg); err != nil {
				errChan <- err
				break
			}
		}
		// close send stream
		if err := stream.CloseSend(); err != nil {
			errChan <- err
		}
	}(errChan)
	// receive resp
	var resp []*api_v2.StreamingMessagePong
	for {
		select {
		case err := <-errChan:
			return nil, err
		default:
			reply, err := stream.Recv()
			if err == io.EOF {
				return resp, nil
			}
			if err != nil {
				return nil, err
			}
			resp = append(resp, reply)
		}
	}
}

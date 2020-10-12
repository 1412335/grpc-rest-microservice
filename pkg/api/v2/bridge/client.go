package bridge

import (
	"context"
	api_v2 "grpc-rest-microservice/pkg/api/v2/gen/grpc-gateway/gen"
	"io"
	"log"
	"time"

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

type ClientImpl struct {
	client   api_v2.ServiceExtraClient
	conn     *grpcpool.ClientConn
	ctx      context.Context
	username string
	password string
}

func (c *ClientImpl) Close() error {
	return c.conn.Close()
}

func (c *ClientImpl) setHeader(m map[string]string) (context.Context, error) {
	md := metadata.New(m)
	ctx := metadata.NewOutgoingContext(c.ctx, md)
	return ctx, nil
}

// login & get token
func (c *ClientImpl) Login() (string, error) {
	ctx, err := c.setHeader(map[string]string{"custom-req-header": "login"})
	if err != nil {
		return "", err
	}

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
func (c *ClientImpl) Ping(timestamp int64) (*api_v2.MessagePong, error) {
	ctx, err := c.setHeader(map[string]string{"custom-req-header": "ping"})
	if err != nil {
		return nil, err
	}

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

// server streaming
func (c *ClientImpl) StreamingPing(timestamp int64, count, interval int32) ([]*api_v2.StreamingMessagePong, error) {

	ctx, err := c.setHeader(map[string]string{"custom-req-header": "server-streaming-ping"})
	if err != nil {
		return nil, err
	}

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
func (c *ClientImpl) StreamingPost(in []*api_v2.StreamingMessagePing) (*api_v2.StreamingMessagePong, error) {

	ctx, err := c.setHeader(map[string]string{"custom-req-header": "client-streaming-post"})
	if err != nil {
		return nil, err
	}

	stream, err := c.client.StreamingPost(ctx)
	if err != nil {
		return nil, err
	}
	// send msg into stream
	for _, msg := range in {
		if err := stream.Send(msg); err != nil {
			return nil, err
		}
	}
	// close send stream
	if err := stream.CloseSend(); err != nil {
		return nil, err
	}
	// receive resp
	reply, err := stream.CloseAndRecv()
	if err != nil {
		return nil, err
	}
	return reply, nil
}

// bi-directional streaming
func (c *ClientImpl) DuplexStreaming(in []*api_v2.StreamingMessagePing) ([]*api_v2.StreamingMessagePong, error) {
	ctx, err := c.setHeader(map[string]string{"custom-req-header": "duplex-streaming"})
	if err != nil {
		return nil, err
	}

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

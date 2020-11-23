package v2

import (
	"context"
	"log"
	"strings"

	"google.golang.org/grpc"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"google.golang.org/grpc/metadata"

	api_v2 "github.com/1412335/grpc-rest-microservice/pkg/api/v2/gen/grpc-gateway/gen"
)

const (
	apiVersion = "v2"
)

type ServiceAImpl struct{}

func NewServiceAImpl() api_v2.ServiceAServer {
	return &ServiceAImpl{}
}

func (r *ServiceAImpl) Ping(ctx context.Context, req *api_v2.MessagePing) (*api_v2.MessagePong, error) {
	log.Println("a-ping req:", req)

	// Anything linked to this variable will fetch request headers.
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.DataLoss, "failed to get metadata")
	}
	xrid := md.Get("x-request-id")
	if len(xrid) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "missing 'x-request-id' header")
	}
	if strings.Trim(xrid[0], " ") == "" {
		return nil, status.Errorf(codes.InvalidArgument, "empty 'x-request-id' header")
	}

	log.Println("a-ping x-request-id:", xrid[0])

	// Anything linked to this variable will transmit response headers.
	header := metadata.New(map[string]string{"x-response-id": "resp-ping-a"})
	if err := grpc.SendHeader(ctx, header); err != nil {
		return nil, status.Errorf(codes.Internal, "unable to send 'x-response-id' header")
	}

	return &api_v2.MessagePong{
		Timestamp:   req.GetTimestamp(),
		ServiceName: "ServiceA: ping",
	}, nil
}

func (r *ServiceAImpl) Post(ctx context.Context, req *api_v2.MessagePing) (*api_v2.MessagePong, error) {
	log.Println("a-post req:", req)

	// fetch request headers
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.DataLoss, "failed to get metadata")
	}
	xrid := md.Get("x-request-id")
	if len(xrid) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "missing 'x-request-id' header")
	}
	if strings.Trim(xrid[0], " ") == "" {
		return nil, status.Errorf(codes.InvalidArgument, "empty 'x-request-id' header")
	}
	log.Println("a-post x-request-id:", xrid[0])

	// Anything linked to this variable will transmit response headers.
	header := metadata.New(map[string]string{"x-response-id": "resp-post-a"})
	if err := grpc.SendHeader(ctx, header); err != nil {
		return nil, status.Errorf(codes.Internal, "unable to send 'x-response-id' header")
	}

	return &api_v2.MessagePong{
		Timestamp:   req.GetTimestamp(),
		ServiceName: "ServiceA: post",
	}, nil
}

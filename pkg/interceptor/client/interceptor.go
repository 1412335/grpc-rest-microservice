package interceptor

import (
	"context"

	"google.golang.org/grpc"
)

type ClientInterceptor interface {
	// client interceptor for unary request
	Unary() grpc.UnaryClientInterceptor
	unaryClientInterceptor(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error

	// client streaming interceptor
	Stream() grpc.StreamClientInterceptor
	streamClientInterceptor(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error)
}

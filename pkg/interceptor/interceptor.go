package interceptor

import (
	"context"

	"google.golang.org/grpc"
)

type ServerInterceptor interface {
	// unary request to grpc server
	Unary() grpc.UnaryServerInterceptor
	unaryServerInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error)

	// stream request interceptor
	Stream() grpc.StreamServerInterceptor
	streamServerInterceptor(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error)
}

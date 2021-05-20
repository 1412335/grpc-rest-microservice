package proxy

import (
	"context"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"google.golang.org/grpc"
)

type RegisterServiceHandler func(ctx context.Context, mux *runtime.ServeMux, endpoint string, opts []grpc.DialOption) (err error)

type Proxy interface {
	Run() error
	RegisterServiceHandlerFromEndpoint([]RegisterServiceHandler) error
}

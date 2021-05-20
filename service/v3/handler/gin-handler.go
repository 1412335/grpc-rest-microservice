package handler

import (
	_ "github.com/gogo/gateway"

	api_v3 "github.com/1412335/grpc-rest-microservice/pkg/api/v3"
	"github.com/1412335/grpc-rest-microservice/pkg/configs"
	"github.com/1412335/grpc-rest-microservice/pkg/log"
	"github.com/1412335/grpc-rest-microservice/pkg/proxy"

	// Static files
	_ "github.com/1412335/grpc-rest-microservice/pkg/api/v3/statik"
)

func NewProxy(config *configs.ServiceConfig) proxy.Proxy {
	return proxy.NewHandler(config, log.DefaultLogger, []proxy.RegisterServiceHandler{
		api_v3.RegisterUserServiceHandlerFromEndpoint,
	})
}

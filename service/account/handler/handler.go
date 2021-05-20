package handler

import (
	pb "account/api"

	"github.com/1412335/grpc-rest-microservice/pkg/configs"
	"github.com/1412335/grpc-rest-microservice/pkg/log"
	"github.com/1412335/grpc-rest-microservice/pkg/proxy"

	// Static files
	_ "account/api/statik"
)

func NewHandler(config *configs.ServiceConfig) proxy.Proxy {
	return proxy.NewHandler(config, log.DefaultLogger, []proxy.RegisterServiceHandler{
		pb.RegisterAccountServiceHandlerFromEndpoint,
	})
}

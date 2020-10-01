package v2

import (
	"context"
	"log"

	api_v2 "grpc-rest-microservice/pkg/api/v2/gen/grpc-gateway/gen"
)

type ServiceExtraImpl struct{}

func NewServiceExtraImpl() api_v2.ServiceExtraServer {
	return &ServiceExtraImpl{}
}

func (r *ServiceExtraImpl) Ping(ctx context.Context, req *api_v2.MessagePing) (*api_v2.MessagePong, error) {
	log.Println("extra", req)
	return &api_v2.MessagePong{
		Timestamp:   req.GetTimestamp(),
		ServiceName: "ServiceExtra: ping",
	}, nil
}

func (r *ServiceExtraImpl) Post(ctx context.Context, req *api_v2.MessagePing) (*api_v2.MessagePong, error) {
	log.Println("extra", req)
	return &api_v2.MessagePong{
		Timestamp:   req.GetTimestamp(),
		ServiceName: "ServiceExtra: post",
	}, nil
}

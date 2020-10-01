package v2

import (
	"context"
	"log"

	api_v2 "grpc-rest-microservice/pkg/api/v2/gen/grpc-gateway/gen"
)

const (
	apiVersion = "v2"
)

type ServiceAImpl struct{}

func NewServiceAImpl() api_v2.ServiceAServer {
	return &ServiceAImpl{}
}

func (r *ServiceAImpl) Ping(ctx context.Context, req *api_v2.MessagePing) (*api_v2.MessagePong, error) {
	log.Println(req)
	return &api_v2.MessagePong{
		Timestamp:   req.GetTimestamp(),
		ServiceName: "ServiceA: ping",
	}, nil
}

func (r *ServiceAImpl) Post(ctx context.Context, req *api_v2.MessagePing) (*api_v2.MessagePong, error) {
	log.Println(req)
	return &api_v2.MessagePong{
		Timestamp:   req.GetTimestamp(),
		ServiceName: "ServiceA: post",
	}, nil
}

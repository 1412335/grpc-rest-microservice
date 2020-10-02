package v2

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"

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

func (r *ServiceExtraImpl) StreamingPing(req *api_v2.StreamingMessagePing, stream api_v2.ServiceExtra_StreamingPingServer) error {
	count := req.GetMessageCount()
	for i := int32(0); i < count; i++ {
		reply := &api_v2.StreamingMessagePong{
			Timestamp:   req.GetTimestamp(),
			ServiceName: fmt.Sprintf("ServiceExtra: streaming ping => %d", i),
		}
		err := stream.Send(reply)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *ServiceExtraImpl) StreamingPost(stream api_v2.ServiceExtra_StreamingPostServer) error {
	var count int32
	startTime := time.Now()
	for {
		_, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				elapsedTime := int64(time.Now().Sub(startTime).Seconds())
				return stream.SendAndClose(&api_v2.StreamingMessagePong{
					Timestamp:   elapsedTime,
					ServiceName: fmt.Sprintf("ServiceExtra: streaming post one-way => %d", count),
				})
			}
			return err
		}
		count++
	}
}

func (r *ServiceExtraImpl) DuplexStreamingPing(stream api_v2.ServiceExtra_DuplexStreamingPingServer) error {
	for {
		in, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}

		count := in.GetMessageCount()
		for i := int32(0); i < count; i++ {
			reply := &api_v2.StreamingMessagePong{
				Timestamp:   in.GetTimestamp(),
				ServiceName: fmt.Sprintf("ServiceExtra: streaming ping => %d", i),
			}
			if err := stream.Send(reply); err != nil {
				return err
			}
		}
	}
}

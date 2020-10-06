package v2

import (
	"context"
	"fmt"
	"io"
	"log"
	"strings"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	api_v2 "grpc-rest-microservice/pkg/api/v2/gen/grpc-gateway/gen"

	"google.golang.org/grpc/metadata"
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

	// fetch request header
	md, ok := metadata.FromIncomingContext(stream.Context())
	if !ok {
		return status.Errorf(codes.DataLoss, "failed to get metadata")
	}
	xrid := md.Get("x-request-id")
	if len(xrid) == 0 {
		return status.Errorf(codes.InvalidArgument, "missing 'x-request-id' header")
	}
	if strings.Trim(xrid[0], " ") == "" {
		return status.Errorf(codes.InvalidArgument, "empty 'x-request-id' header")
	}
	log.Println("stream-ping x-request-id", xrid[0])

	count := req.GetMessageCount()

	// send header response
	if err := stream.SendHeader(metadata.New(map[string]string{
		"x-response-id": fmt.Sprintf("%d", count),
	})); err != nil {
		return status.Errorf(codes.Internal, "unable to send response 'x-response-id' header: %v", err)
	}

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

	stream.SetTrailer(metadata.New(map[string]string{
		"foo": "foo2",
		"bar": "bar2",
	}))

	return nil
}

func (r *ServiceExtraImpl) StreamingPost(stream api_v2.ServiceExtra_StreamingPostServer) error {
	var count int32
	startTime := time.Now()

	ctx := stream.Context()

	// receive header from request
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return status.Errorf(codes.DataLoss, "failed to get metadata")
	}
	xrid := md.Get("x-request-id")
	if len(xrid) == 0 {
		return status.Errorf(codes.InvalidArgument, "missing 'x-request-id' header")
	}
	if strings.Trim(xrid[0], " ") == "" {
		return status.Errorf(codes.InvalidArgument, "empty 'x-request-id' header")
	}
	log.Println("stream-post x-request-id", xrid[0])

	for {
		msg, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				elapsedTime := int64(time.Now().Sub(startTime).Seconds())

				// send header response
				if err := stream.SendHeader(metadata.New(map[string]string{
					"count": fmt.Sprintf("%d", count),
				})); err != nil {
					return status.Errorf(codes.Internal, "unable to send response header count")
				}

				stream.SetTrailer(metadata.New(map[string]string{
					"foo": "foo2",
					"bar": "bar2",
				}))
				return stream.SendAndClose(&api_v2.StreamingMessagePong{
					Timestamp:   elapsedTime,
					ServiceName: fmt.Sprintf("ServiceExtra: streaming post one-way => %d", count),
				})
			}
			return status.Errorf(codes.Internal, "received stream failed: %v", err)
		}
		log.Println("stream msg", msg)

		count++
	}
}

func (r *ServiceExtraImpl) DuplexStreamingPing(stream api_v2.ServiceExtra_DuplexStreamingPingServer) error {

	ctx := stream.Context()

	// receive header from request
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return status.Errorf(codes.DataLoss, "failed to get metadata")
	}
	xrid := md.Get("x-request-id")
	if len(xrid) == 0 {
		return status.Errorf(codes.InvalidArgument, "missing 'x-request-id' header")
	}
	if strings.Trim(xrid[0], " ") == "" {
		return status.Errorf(codes.InvalidArgument, "empty 'x-request-id' header")
	}
	log.Println("stream-duplex x-request-id", xrid[0])

	// send header response
	if err := stream.SendHeader(metadata.New(map[string]string{
		"x-response-id": "duplex-stream",
	})); err != nil {
		return status.Errorf(codes.Internal, "unable to send 'x-response-id' header: %v", err)
	}

	for {
		in, err := stream.Recv()
		if err != nil {
			if err == io.EOF {

				stream.SetTrailer(metadata.New(map[string]string{
					"foo": "foo2",
					"bar": "bar2",
				}))

				return nil
			}
			return status.Errorf(codes.Internal, "stream received failed: %v", err)
		}

		count := in.GetMessageCount()

		for i := int32(0); i < count; i++ {
			reply := &api_v2.StreamingMessagePong{
				Timestamp:   in.GetTimestamp(),
				ServiceName: fmt.Sprintf("ServiceExtra: streaming ping => %d", i),
			}

			if err := stream.Send(reply); err != nil {
				return status.Errorf(codes.Internal, "stream send failed: %v", err)
			}
		}
	}
}

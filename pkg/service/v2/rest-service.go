package v2

import (
	"context"
	"strings"

	"go.uber.org/zap"
	"google.golang.org/grpc"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"google.golang.org/grpc/metadata"

	api_v2 "github.com/1412335/grpc-rest-microservice/pkg/api/v2/gen/grpc-gateway/gen"
	"github.com/1412335/grpc-rest-microservice/pkg/log"
)

type serviceAImpl struct {
	logger log.Factory
}

func NewServiceAImpl(logger log.Factory) api_v2.ServiceAServer {
	return &serviceAImpl{
		logger: logger,
	}
}

func (s *serviceAImpl) getRequestID(ctx context.Context) (string, error) {
	// Anything linked to this variable will fetch request headers.
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", status.Errorf(codes.DataLoss, "failed to get metadata")
	}
	xrid := md.Get("x-request-id")
	if len(xrid) == 0 {
		return "", status.Errorf(codes.InvalidArgument, "missing 'x-request-id' header")
	}
	if strings.Trim(xrid[0], " ") == "" {
		return "", status.Errorf(codes.InvalidArgument, "empty 'x-request-id' header")
	}
	return xrid[0], nil
}

func (s *serviceAImpl) sendHeaderResp(ctx context.Context, header metadata.MD) error {
	if err := grpc.SendHeader(ctx, header); err != nil {
		return status.Errorf(codes.Internal, "unable to send 'x-response-id' header")
	}
	return nil
}

func (s *serviceAImpl) unaryRequest(ctx context.Context, req *api_v2.MessagePing, method string) (*api_v2.MessagePong, error) {
	s.logger.For(ctx).Info("unary request", zap.String("method", method), zap.Any("req", req))

	xrid, err := s.getRequestID(ctx)
	if err != nil {
		return nil, err
	}
	s.logger.For(ctx).Info("request id", zap.String("x-request-id", xrid))

	// Anything linked to this variable will transmit response headers.
	header := metadata.New(map[string]string{"x-response-id": method})
	if err := s.sendHeaderResp(ctx, header); err != nil {
		return nil, status.Errorf(codes.Internal, "unable to send 'x-response-id' header")
	}

	return &api_v2.MessagePong{
		Timestamp:   req.GetTimestamp(),
		ServiceName: "ServiceA: " + method,
	}, nil
}

func (s *serviceAImpl) Ping(ctx context.Context, req *api_v2.MessagePing) (*api_v2.MessagePong, error) {
	return s.unaryRequest(ctx, req, "ping")
}

func (s *serviceAImpl) Post(ctx context.Context, req *api_v2.MessagePing) (*api_v2.MessagePong, error) {
	return s.unaryRequest(ctx, req, "post")
}

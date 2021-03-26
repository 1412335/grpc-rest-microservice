package interceptor

import (
	"context"
	"fmt"
	"strings"

	api_v2 "github.com/1412335/grpc-rest-microservice/pkg/api/v2/grpc-gateway/gen"
	"github.com/1412335/grpc-rest-microservice/pkg/log"
	"go.uber.org/zap"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// Simple interceptor
type SimpleServerInterceptor struct {
	logger log.Factory
}

func (interceptor *SimpleServerInterceptor) WithLogger(logger log.Factory) {
	interceptor.logger = logger
}

func (interceptor *SimpleServerInterceptor) Unary() grpc.UnaryServerInterceptor {
	return interceptor.unaryServerInterceptor
}

func (interceptor *SimpleServerInterceptor) Stream() grpc.StreamServerInterceptor {
	return interceptor.streamServerInterceptor
}

// unary request to grpc server
func (interceptor *SimpleServerInterceptor) unaryServerInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	var xrid []string
	var customHeader []string
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic: %v", r)
		}
	}()
	defer func() {
		interceptor.logger.For(ctx).Info("unary request", zap.String("method", info.FullMethod), zap.Any("customHeader", customHeader), zap.Any("xrid", xrid), zap.Error(err))
	}()

	// fetch headers req
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.DataLoss, "failed to get metadata")
	}
	xrid = md.Get("x-request-id")
	if len(xrid) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "missing 'x-request-id' header")
	}
	if strings.Trim(xrid[0], " ") == "" {
		return nil, status.Errorf(codes.InvalidArgument, "empty 'x-request-id' header")
	}

	// fetch custom-request-header
	customHeader = md.Get("custom-req-header")

	// validate request
	// log.Println("[gRPC server] validate req")

	// send x-response-id header
	header := metadata.New(map[string]string{"x-response-id": xrid[0]})
	if e := grpc.SetHeader(ctx, header); e != nil {
		return nil, status.Errorf(codes.Internal, "unable to send 'x-response-id' header: %v", e.Error())
	}
	// NOT WORK: because server service does NOT using context to send anything
	// ctx = metadata.AppendToOutgoingContext(ctx, []string{"x-response-id", xrid[0]}...)

	// check request timeout or canceled by the client
	if ctx.Err() == context.Canceled {
		return nil, status.Error(codes.Canceled, "request is canceled")
	}
	if ctx.Err() == context.DeadlineExceeded {
		return nil, status.Error(codes.DeadlineExceeded, "deadline is exceeded")
	}

	resp, err = handler(ctx, req)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "serve handler error: %+v", err)
	}

	// add serviceName into response
	if msg, ok := resp.(*api_v2.MessagePong); ok {
		msg.ServiceName = info.FullMethod
		return msg, nil
	}

	return resp, nil
}

// stream request interceptor
func (interceptor *SimpleServerInterceptor) streamServerInterceptor(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic: %v", r)
		}
	}()

	// fetch x-request-id header
	md, ok := metadata.FromIncomingContext(ss.Context())
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

	// fetch custom-request-header
	customHeader := md.Get("custom-req-header")

	interceptor.logger.For(ss.Context()).Info("stream request",
		zap.String("method", info.FullMethod),
		zap.Any("customHeader", customHeader),
		zap.Any("xrid", xrid),
		zap.Bool("serverStream", info.IsServerStream),
		zap.Error(err),
	)

	// send x-response-id header
	header := metadata.New(map[string]string{
		"x-response-id": xrid[0],
	})
	if err := ss.SendHeader(header); err != nil {
		return status.Errorf(codes.Internal, "unable to send response 'x-response-id' header: %v", err)
	}

	return handler(srv, ss)
}

package interceptor

import (
	"context"
	"fmt"
	api_v2 "grpc-rest-microservice/pkg/api/v2/gen/grpc-gateway/gen"
	"grpc-rest-microservice/pkg/configs"
	"log"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// Credentials interceptor
type CredentialsServerInterceptor struct {
	username string
	password string
}

func NewCredentialsServerInterceptor(config *configs.Authentication) *CredentialsServerInterceptor {
	return &CredentialsServerInterceptor{
		username: config.Username,
		password: config.Password,
	}
}

func (interceptor *CredentialsServerInterceptor) Unary() grpc.UnaryServerInterceptor {
	return interceptor.unaryServerInterceptor
}

func (interceptor *CredentialsServerInterceptor) Stream() grpc.StreamServerInterceptor {
	return interceptor.streamServerInterceptor
}

// unary request to grpc server
func (interceptor *CredentialsServerInterceptor) unaryServerInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {

	var xrid []string
	var customHeader []string
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic: %v", r)
		}
	}()
	defer func() {
		log.Printf("[gRPC server][credendtials] Received RPC method=%s, xrid=%v, customHeader=%v, error='%v'", info.FullMethod, xrid, customHeader, err)
	}()

	// fetch headers req
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.DataLoss, "failed to get metadata")
	}
	// AUTHENTICATION WITH CREDENTIALS ----
	var (
		username string
		password string
	)
	if val, ok := md["username"]; ok {
		username = strings.Trim(val[0], " ")
	}
	if val, ok := md["password"]; ok {
		password = strings.Trim(val[0], " ")
	}
	if username == "" || password == "" {
		return nil, status.Errorf(codes.InvalidArgument, "missing credentials metadata")
	}
	// validate username and password
	if username != interceptor.username || password != interceptor.password {
		return nil, status.Errorf(codes.Unauthenticated, "invalid credentials [username:%s, password:%s]", username, password)
	}
	// END AUTHENTICATION WITH CREDENTIALS ----

	// get x-request-id header
	xrid = md.Get("x-request-id")
	if len(xrid) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "missing 'x-request-id' header")
	}
	if strings.Trim(xrid[0], " ") == "" {
		return nil, status.Errorf(codes.InvalidArgument, "empty 'x-request-id' header")
	}

	// send x-response-id header
	header := metadata.New(map[string]string{"x-response-id": xrid[0]})
	if err := grpc.SetHeader(ctx, header); err != nil {
		return nil, status.Errorf(codes.Internal, "unable to send 'x-response-id' header")
	}
	// NOT WORK: because server service does NOT using context to send anything
	// ctx = metadata.AppendToOutgoingContext(ctx, []string{"x-response-id", xrid[0]}...)

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
func (interceptor *CredentialsServerInterceptor) streamServerInterceptor(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
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

	// AUTHENTICATION WITH CREDENTIALS ----
	var (
		username string
		password string
	)
	if val, ok := md["username"]; ok {
		username = strings.Trim(val[0], " ")
	}
	if val, ok := md["password"]; ok {
		password = strings.Trim(val[0], " ")
	}
	if username == "" || password == "" {
		return status.Errorf(codes.InvalidArgument, "missing credentials metadata")
	}
	// validate username and password
	if username != interceptor.username || password != interceptor.password {
		return status.Errorf(codes.Unauthenticated, "invalid credentials [username:%s, password:%s]", username, password)
	}
	// END AUTHENTICATION WITH CREDENTIALS ----

	xrid := md.Get("x-request-id")
	if len(xrid) == 0 {
		return status.Errorf(codes.InvalidArgument, "missing 'x-request-id' header")
	}
	if strings.Trim(xrid[0], " ") == "" {
		return status.Errorf(codes.InvalidArgument, "empty 'x-request-id' header")
	}

	// fetch custom-request-header
	customHeader := md.Get("custom-req-header")

	log.Printf("[gRPC server][credendtials] Received Stream RPC method=%s, serverStream=%v, xrid=%v, customHeader=%v, error='%v'", info.FullMethod, info.IsServerStream, xrid, customHeader, err)

	// send x-response-id header
	header := metadata.New(map[string]string{
		"x-response-id": xrid[0],
	})
	if err := ss.SendHeader(header); err != nil {
		return status.Errorf(codes.Internal, "unable to send response 'x-response-id' header: %v", err)
	}

	return handler(srv, ss)
}

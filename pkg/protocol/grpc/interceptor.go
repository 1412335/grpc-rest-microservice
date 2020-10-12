package grpc

import (
	"context"
	"fmt"
	api_v2 "grpc-rest-microservice/pkg/api/v2/gen/grpc-gateway/gen"
	"grpc-rest-microservice/pkg/utils"
	"log"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type ServerInterceptor interface {
	// unary request to grpc server
	Unary() grpc.UnaryServerInterceptor
	unaryServerInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error)

	// stream request interceptor
	Stream() grpc.StreamServerInterceptor
	streamServerInterceptor(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error)
}

// Simple interceptor
type SimpleServerInterceptor struct{}

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
		log.Printf("[gRPC server] Received RPC method=%s, xrid=%v, customHeader=%v, error='%v'", info.FullMethod, xrid, customHeader, err)
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

	log.Printf("[gRPC server] Received Stream RPC method=%s, serverStream=%v, xrid=%v, customHeader=%v, error='%v'", info.FullMethod, info.IsServerStream, xrid, customHeader, err)

	// send x-response-id header
	header := metadata.New(map[string]string{
		"x-response-id": xrid[0],
	})
	if err := ss.SendHeader(header); err != nil {
		return status.Errorf(codes.Internal, "unable to send response 'x-response-id' header: %v", err)
	}

	return handler(srv, ss)
}

// Auth interceptor with JWT
type AuthServerInterceptor struct {
	jwtManager      *utils.JWTManager
	accessibleRoles map[string][]string
}

func NewAuthServerInterceptor(jwtManager *utils.JWTManager, accessibleRoles map[string][]string) *AuthServerInterceptor {
	return &AuthServerInterceptor{jwtManager, accessibleRoles}
}

func (interceptor *AuthServerInterceptor) Unary() grpc.UnaryServerInterceptor {
	return interceptor.unaryServerInterceptor
}

func (interceptor *AuthServerInterceptor) Stream() grpc.StreamServerInterceptor {
	return interceptor.streamServerInterceptor
}

func (interceptor *AuthServerInterceptor) authorize(ctx context.Context, method string) error {
	accessibleRoles, ok := interceptor.accessibleRoles[method]
	log.Printf("check %+v '%s' %v %v\n", interceptor.accessibleRoles, method, accessibleRoles, ok)
	if !ok {
		return nil
	}

	// fetch authorization header
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return status.Errorf(codes.DataLoss, "failed to get metadata")
	}
	accessToken := md.Get("authorization")
	log.Println("accessToken", accessToken)
	if len(accessToken) == 0 {
		return status.Errorf(codes.InvalidArgument, "missing 'authorization' header")
	}
	if strings.Trim(accessToken[0], " ") == "" {
		return status.Errorf(codes.InvalidArgument, "empty 'authorization' header")
	}

	userClaims, err := interceptor.jwtManager.Verify(accessToken[0])
	if err != nil {
		return status.Errorf(codes.Unauthenticated, "verify failed: %v", err)
	}
	for _, role := range accessibleRoles {
		if role == userClaims.Role {
			return nil
		}
	}
	// fetch custom-request-header
	// customHeader = md.Get("custom-req-header")

	// validate request
	// log.Println("[gRPC server] validate req")
	return status.Errorf(codes.PermissionDenied, "no permission to access this method: %s with [username:%s, role:%s]", method, userClaims.Username, userClaims.Role)
}

// unary request to grpc server
func (interceptor *AuthServerInterceptor) unaryServerInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic: %v", r)
		}
	}()
	log.Printf("[gRPC server] Received RPC method=%s", info.FullMethod)

	err = interceptor.authorize(ctx, info.FullMethod)
	if err != nil {
		return nil, err
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
func (interceptor *AuthServerInterceptor) streamServerInterceptor(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic: %v", r)
		}
	}()
	log.Printf("[gRPC server] Received Stream RPC method=%s, serverStream=%v", info.FullMethod, info.IsServerStream)

	err = interceptor.authorize(ss.Context(), info.FullMethod)
	if err != nil {
		return err
	}

	// send x-response-id header
	header := metadata.New(map[string]string{
		"x-response-id": "interceptor-streaming",
	})
	if err := ss.SendHeader(header); err != nil {
		return status.Errorf(codes.Internal, "unable to send response 'x-response-id' header: %v", err)
	}

	return handler(srv, ss)
}

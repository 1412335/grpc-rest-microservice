package interceptor

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
	log.Printf("[gRPC server][auth] Received RPC method=%s", info.FullMethod)

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
	log.Printf("[gRPC server][auth] Received Stream RPC method=%s, serverStream=%v", info.FullMethod, info.IsServerStream)

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
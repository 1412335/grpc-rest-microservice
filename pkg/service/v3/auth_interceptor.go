package v3

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

// Auth interceptor with JWT
type AuthServerInterceptor struct {
	logger          log.Factory
	jwtManager      *TokenService
	accessibleRoles map[string][]string
}

func NewAuthServerInterceptor(logger log.Factory, jwtManager *TokenService, accessibleRoles map[string][]string) *AuthServerInterceptor {
	return &AuthServerInterceptor{logger, jwtManager, accessibleRoles}
}

func (a *AuthServerInterceptor) Unary() grpc.UnaryServerInterceptor {
	return a.unaryServerInterceptor
}

func (a *AuthServerInterceptor) Stream() grpc.StreamServerInterceptor {
	return a.streamServerInterceptor
}

func (a *AuthServerInterceptor) authorize(ctx context.Context, method string) error {
	accessibleRoles, ok := a.accessibleRoles[method]
	a.logger.For(ctx).Info("authorize", zap.String("method", method), zap.Any("accessibleRoles", accessibleRoles), zap.Bool("ok", ok))
	if !ok {
		return nil
	}

	// fetch authorization header
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return status.Errorf(codes.DataLoss, "failed to get metadata")
	}
	accessToken := md.Get("authorization")
	if len(accessToken) == 0 {
		return status.Errorf(codes.InvalidArgument, "missing 'authorization' header")
	}
	if strings.Trim(accessToken[0], " ") == "" {
		return status.Errorf(codes.InvalidArgument, "empty 'authorization' header")
	}
	a.logger.For(ctx).Info("accessToken", zap.String("accessToken", accessToken[0]))

	userClaims, err := a.jwtManager.Verify(accessToken[0])
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
func (a *AuthServerInterceptor) unaryServerInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic: %v", r)
		}
	}()
	a.logger.For(ctx).Info("unary req", zap.String("method", info.FullMethod))

	err = a.authorize(ctx, info.FullMethod)
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
func (a *AuthServerInterceptor) streamServerInterceptor(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic: %v", r)
		}
	}()
	a.logger.For(ss.Context()).Info("stream req", zap.String("method", info.FullMethod), zap.Any("serverStream", info.IsServerStream))

	err = a.authorize(ss.Context(), info.FullMethod)
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

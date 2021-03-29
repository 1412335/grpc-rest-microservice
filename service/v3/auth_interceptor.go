package v3

import (
	"context"
	"strings"

	api_v3 "github.com/1412335/grpc-rest-microservice/pkg/api/v3"
	interceptor "github.com/1412335/grpc-rest-microservice/pkg/interceptor/server"
	"github.com/1412335/grpc-rest-microservice/pkg/log"
	"go.uber.org/zap"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// Auth interceptor with JWT
type AuthServerInterceptor struct {
	interceptor.ServerInterceptor
	logger          log.Factory
	jwtManager      *TokenService
	accessibleRoles map[string][]string
}

var _ interceptor.ServerInterceptor = (*AuthServerInterceptor)(nil)

func NewAuthServerInterceptor(logger log.Factory, jwtManager *TokenService, accessibleRoles map[string][]string) *AuthServerInterceptor {
	return &AuthServerInterceptor{
		logger:          logger,
		jwtManager:      jwtManager,
		accessibleRoles: accessibleRoles,
	}
}

func (a *AuthServerInterceptor) Unary() grpc.UnaryServerInterceptor {
	return a.unaryServerInterceptor
}

func (a *AuthServerInterceptor) Stream() grpc.StreamServerInterceptor {
	return a.streamServerInterceptor
}

func (a *AuthServerInterceptor) authorize(ctx context.Context, method string, req interface{}) error {
	// check accessiable method with user role got from header authorization
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

	// verify token
	userClaims, err := a.jwtManager.Verify(accessToken[0])
	if err != nil {
		return status.Errorf(codes.Unauthenticated, "verify failed: %v", err)
	}

	// check user self update
	if msg, ok := req.(*api_v3.UpdateUserRequest); ok && msg.GetUser().GetId() == userClaims.ID {
		return nil
	}

	// check role
	for _, role := range accessibleRoles {
		if role == api_v3.Role_ROOT.String() || role == strings.ToLower(userClaims.Role) {
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
			a.logger.For(ctx).Error("unary req", zap.Any("panic", r))
			err = status.Error(codes.Unknown, "server error")
		}
	}()
	a.logger.For(ctx).Info("unary req", zap.String("method", info.FullMethod))

	// authorize request
	err = a.authorize(ctx, info.FullMethod, req)
	if err != nil {
		return nil, err
	}

	// NOT WORK: because server service does NOT using context to send anything
	ctx = metadata.AppendToOutgoingContext(ctx, []string{"x-response-id", "a"}...)

	// add serviceName into response
	// if msg, ok := req.(*api_v3.UpdateUserRequest); ok {
	// 	msg. = info.FullMethod
	// 	return msg, nil
	// }

	return handler(ctx, req)
}

// stream request interceptor
func (a *AuthServerInterceptor) streamServerInterceptor(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
	defer func() {
		if r := recover(); r != nil {
			a.logger.For(ss.Context()).Error("stream req", zap.Any("panic", r))
			err = status.Error(codes.Unknown, "server error")
		}
	}()
	a.logger.For(ss.Context()).Info("stream req", zap.String("method", info.FullMethod), zap.Any("serverStream", info.IsServerStream))

	err = a.authorize(ss.Context(), info.FullMethod, nil)
	if err != nil {
		return err
	}

	// send x-response-id header
	header := metadata.New(map[string]string{
		"x-response-id": "auth-streaming",
	})
	if err = ss.SendHeader(header); err != nil {
		return status.Errorf(codes.Unknown, "unable to send response 'x-response-id' header: %v", err)
	}

	err = handler(srv, ss)
	if err != nil {
		return err
	}

	// return error when metadata includes error header
	if header, ok := metadata.FromIncomingContext(ss.Context()); ok {
		if v, ok := header["error"]; ok {
			ss.SetTrailer(metadata.New(map[string]string{
				"foo": "foo2",
				"bar": "bar2",
			}))
			return status.Errorf(codes.InvalidArgument, "error metadata: %v", v)
		}
	}
	return nil
}

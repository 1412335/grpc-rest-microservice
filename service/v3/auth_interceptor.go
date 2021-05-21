package v3

import (
	"context"
	"strings"

	api_v3 "github.com/1412335/grpc-rest-microservice/pkg/api/v3"
	interceptor "github.com/1412335/grpc-rest-microservice/pkg/interceptor/server"
	"github.com/1412335/grpc-rest-microservice/pkg/log"
	"github.com/1412335/grpc-rest-microservice/pkg/utils"

	"go.uber.org/zap"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// Auth interceptor with JWT
type AuthServerInterceptor struct {
	jwtManager          *TokenService
	authRequiredMethods map[string]bool
	accessibleRoles     map[string][]string
}

var _ interceptor.ServerInterceptor = (*AuthServerInterceptor)(nil)

func NewAuthServerInterceptor(jwtManager *TokenService, authRequiredMethods map[string]bool, accessibleRoles map[string][]string) *AuthServerInterceptor {
	return &AuthServerInterceptor{
		jwtManager:          jwtManager,
		authRequiredMethods: authRequiredMethods,
		accessibleRoles:     accessibleRoles,
	}
}

func (a *AuthServerInterceptor) Log() log.Factory {
	return interceptor.DefaultLogger.With(zap.String("interceptor-name", "auth"))
}

func (a *AuthServerInterceptor) Unary() grpc.UnaryServerInterceptor {
	return a.UnaryInterceptor
}
func (a *AuthServerInterceptor) Stream() grpc.StreamServerInterceptor {
	return a.StreamInterceptor
}

func (a *AuthServerInterceptor) authorize(ctx context.Context, method string, req interface{}) error {
	authReq, ok := a.authRequiredMethods[method]
	if !authReq || !ok {
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
	a.Log().For(ctx).Info("authorize", zap.String("token", accessToken[0]))

	// verify token
	userClaims, err := a.jwtManager.Verify(accessToken[0])
	if err != nil {
		return status.Errorf(codes.Unauthenticated, "verify failed: %v", err)
	}

	// invalidate token
	if invalidate, _ := a.jwtManager.IsInvalidated(userClaims.ID, userClaims.Id); invalidate {
		return status.Errorf(codes.Unauthenticated, "invalidated token")
	}

	// root full access
	if strings.ToLower(userClaims.Role) == api_v3.Role_ROOT.String() {
		return nil
	}

	// check accessiable method with user role got from header authorization
	accessibleRoles, ok := a.accessibleRoles[method]
	if !ok {
		return nil
	}
	// check accessible role for method
	for _, role := range accessibleRoles {
		if role == strings.ToLower(userClaims.Role) {
			return nil
		}
	}

	// check action with same userID
	switch method {
	case "/api_v3.UserService/Update":
		if msg, ok := req.(*api_v3.UpdateUserRequest); ok && msg.GetUser().GetId() == userClaims.ID {
			return nil
		}
	case "/api_v3.UserService/Delete":
		if msg, ok := req.(*api_v3.DeleteUserRequest); ok && msg.GetId() == userClaims.ID {
			return nil
		}
	}

	// fetch custom-request-header
	xreqid := md.Get("x-request-id")
	a.Log().For(ctx).Info("request", zap.String("x-request-id", xreqid[0]))

	ctx = utils.SetContextValue(ctx, "userClaims.ID", userClaims.ID)
	userID, ok := utils.GetContextValue(ctx, "userClaims.ID")
	a.Log().Info("ctx token", zap.String("userID", userID), zap.Bool("ok", ok))

	// validate request
	// log.Println("[gRPC server] validate req")
	return status.Errorf(codes.PermissionDenied, "no permission to access this method: %s with [username:%s, role:%s]", method, userClaims.Username, userClaims.Role)
}

// unary request to grpc server
func (a *AuthServerInterceptor) UnaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	defer func() {
		if r := recover(); r != nil {
			a.Log().For(ctx).Error("unary req", zap.Any("panic", r))
			err = status.Error(codes.Unknown, "Internal server error")
		}
	}()
	a.Log().For(ctx).Info("unary req", zap.String("method", info.FullMethod))

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

	// check request timeout or canceled by the client
	if ctx.Err() == context.Canceled {
		return nil, status.Error(codes.Canceled, "request is canceled")
	}
	if ctx.Err() == context.DeadlineExceeded {
		return nil, status.Error(codes.DeadlineExceeded, "deadline is exceeded")
	}

	return handler(ctx, req)
}

// stream request interceptor
func (a *AuthServerInterceptor) StreamInterceptor(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
	defer func() {
		if r := recover(); r != nil {
			a.Log().For(ss.Context()).Error("stream req", zap.Any("panic", r))
			err = status.Error(codes.Unknown, "Internal server error")
		}
	}()
	a.Log().For(ss.Context()).Info("stream req", zap.String("method", info.FullMethod), zap.Any("serverStream", info.IsServerStream))

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

package server

import (
	"context"
	"strings"
	"time"

	pb "account/api"
	"account/client"

	api_v3 "github.com/1412335/grpc-rest-microservice/pkg/api/v3"
	"github.com/1412335/grpc-rest-microservice/pkg/errors"
	interceptor "github.com/1412335/grpc-rest-microservice/pkg/interceptor/server"
	"github.com/1412335/grpc-rest-microservice/pkg/log"

	"go.uber.org/zap"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

// Auth interceptor with JWT
type AuthServerInterceptor struct {
	name                string
	userSrv             client.UserClient
	authRequiredMethods map[string]bool
	accessibleRoles     map[string][]string
}

var _ interceptor.ServerInterceptor = (*AuthServerInterceptor)(nil)

func NewAuthServerInterceptor(userSrv client.UserClient, authRequiredMethods map[string]bool, accessibleRoles map[string][]string) *AuthServerInterceptor {
	return &AuthServerInterceptor{
		name:                "auth",
		userSrv:             userSrv,
		authRequiredMethods: authRequiredMethods,
		accessibleRoles:     accessibleRoles,
	}
}

func (a *AuthServerInterceptor) Log() log.Factory {
	return interceptor.DefaultLogger.With(zap.String("interceptor-name", a.name))
}

func (a *AuthServerInterceptor) Unary() grpc.UnaryServerInterceptor {
	return a.UnaryInterceptor
}
func (a *AuthServerInterceptor) Stream() grpc.StreamServerInterceptor {
	return a.StreamInterceptor
}

func (a *AuthServerInterceptor) authorize(ctx context.Context, method string) (*client.User, error) {
	authReq, ok := a.authRequiredMethods[method]
	if !authReq || !ok {
		return nil, nil
	}

	// fetch authorization header
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.DataLoss, "failed to get metadata")
	}
	accessToken := md.Get("authorization")
	if len(accessToken) == 0 {
		return nil, errors.BadRequest("missing 'authorization' header", nil)
	}
	if strings.Trim(accessToken[0], " ") == "" {
		return nil, errors.BadRequest("empty 'authorization' header", nil)
	}

	// verify token
	user, err := a.userSrv.Validate(accessToken[0])
	if err != nil || user == nil {
		if st, ok := status.FromError(err); ok {
			return nil, st.Err()
		}
		return nil, errors.Unauthenticated("verify failed", "token", err.Error())
	}

	// fetch custom-request-header
	// customHeader = md.Get("custom-req-header")
	return user, nil
}

// unary request to grpc server
func (a *AuthServerInterceptor) UnaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	start := time.Now()
	defer func() {
		a.Log().For(ctx).Info("unary req", zap.String("method", info.FullMethod), zap.Duration("duration", time.Since(start)))
		if r := recover(); r != nil {
			a.Log().For(ctx).Error("unary req", zap.Any("panic", r))
			err = status.Error(codes.Unknown, "Internal server error")
		}
	}()

	// authorize request
	user, err := a.authorize(ctx, info.FullMethod)
	if err != nil {
		return nil, err
	}

	// check action with same userID & add user_id to request
	switch msg := req.(type) {
	case *pb.CreateAccountRequest:
		msg.UserId = user.ID
	case *pb.DeleteAccountRequest:
		msg.UserId = user.ID
	case *pb.UpdateAccountRequest:
		if msg.Account != nil {
			msg.Account.UserId = user.ID
		} else {
			msg.Account = &pb.Account{
				UserId: user.ID,
			}
		}
	case *pb.ListAccountsRequest:
		if user.Role != api_v3.Role_ROOT.String() {
			msg.UserId = wrapperspb.String(user.ID)
		}
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

	_, err = a.authorize(ss.Context(), info.FullMethod)
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

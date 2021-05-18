package client

import (
	"context"
	"fmt"
	"time"

	interceptor "github.com/1412335/grpc-rest-microservice/pkg/interceptor/client"

	uuid "github.com/satori/go.uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type AuthInterceptor struct {
	interceptor.ClientInterceptor
	client      Client
	authMethods map[string]bool
	accessToken string
	email       string
	password    string
}

var _ interceptor.ClientInterceptor = (*AuthInterceptor)(nil)

func NewAuthInterceptor(client Client, authMethods map[string]bool, refreshDuration time.Duration) interceptor.ClientInterceptor {
	i := &AuthInterceptor{
		client:      client,
		authMethods: authMethods,
	}
	// auto refresh auth-token
	err := i.scheduleRefreshToken(refreshDuration)
	if err != nil {
		return nil
	}
	return i
}

// get token using client ping service
func (i *AuthInterceptor) refreshToken() error {
	token, err := i.client.Login(i.email, i.password)
	if err != nil {
		return err
	}
	i.accessToken = token
	return nil
}

// auto refresh token after duration time
func (i *AuthInterceptor) scheduleRefreshToken(refreshDuration time.Duration) error {
	err := i.refreshToken()
	if err != nil {
		return err
	}
	wait := refreshDuration
	go func() {
		for {
			time.Sleep(wait)
			if err := i.refreshToken(); err != nil {
				wait = time.Second
			} else {
				wait = refreshDuration
			}
		}
	}()
	return nil
}

func (i *AuthInterceptor) Unary() grpc.UnaryClientInterceptor {
	return i.unaryClientInterceptor
}

func (i *AuthInterceptor) Stream() grpc.StreamClientInterceptor {
	return i.streamClientInterceptor
}

func (i *AuthInterceptor) attachToken(ctx context.Context) context.Context {
	// read file saved token
	ctx = metadata.AppendToOutgoingContext(ctx, "authorization", i.accessToken)
	return ctx
}

func (i *AuthInterceptor) unaryClientInterceptor(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic: %v", r)
		}
	}()

	// attach jwt token
	if i.authMethods[method] {
		ctx = i.attachToken(ctx)
	}

	// send x-request-id header
	xrid := uuid.NewV4().String()
	// header := metadata.New(map[string]string{"x-request-id": xrid})
	// APPEND HEADER RESQUEST
	ctx = metadata.AppendToOutgoingContext(ctx, []string{"x-request-id", xrid}...)

	// fetch response header
	var header metadata.MD
	opts = append(opts, grpc.Header(&header))

	// invoke request
	return invoker(ctx, method, req, reply, cc, opts...)
}

func (i *AuthInterceptor) streamClientInterceptor(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (clientStream grpc.ClientStream, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic: %v", r)
		}
	}()

	// send x-request-id header
	xrid := uuid.NewV4().String()
	// header := metadata.New(map[string]string{"x-request-id": xrid})
	// APPEND HEADER RESQUEST
	ctx = metadata.AppendToOutgoingContext(ctx, []string{"x-request-id", xrid}...)

	clientStream, err = streamer(ctx, desc, cc, method, opts...)

	// get x-response-id header
	// NOT WORK: not using stream context
	// md, ok := metadata.FromIncomingContext(clientStream.Context())
	// if !ok {
	// 	return nil, status.Errorf(codes.DataLoss, "failed to get metadata")
	// }
	// xrespid := md.Get("x-response-id")

	return clientStream, err
}

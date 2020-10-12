package bridge

import (
	"context"
	"fmt"
	"log"
	"time"

	uuid "github.com/satori/go.uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type ClientInterceptor interface {
	// client interceptor for unary request
	Unary() grpc.UnaryClientInterceptor
	unaryClientInterceptor(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error

	// client streaming interceptor
	Stream() grpc.StreamClientInterceptor
	streamClientInterceptor(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error)
}

type SimpleClientInterceptor struct {
	client      Client
	authMethods map[string]bool
	accessToken string
}

func NewSimpleClientInterceptor(authMethods map[string]bool) (*SimpleClientInterceptor, error) {
	interceptor := &SimpleClientInterceptor{
		// client:      client,
		authMethods: authMethods,
	}
	// err := interceptor.scheduleRefreshToken(refreshDuration)
	// if err != nil {
	// 	return nil, err
	// }
	return interceptor, nil
}

func (interceptor *SimpleClientInterceptor) Load(client Client, authMethods map[string]bool, refreshDuration time.Duration) error {
	interceptor.client = client
	interceptor.authMethods = authMethods
	// load token
	return interceptor.scheduleRefreshToken(refreshDuration)
}

// get token using client ping service
func (interceptor *SimpleClientInterceptor) refreshToken() error {
	token, err := interceptor.client.Login()
	if err != nil {
		return err
	}
	interceptor.accessToken = token
	return nil
}

// auto refresh token after duration time
func (interceptor *SimpleClientInterceptor) scheduleRefreshToken(refreshDuration time.Duration) error {
	err := interceptor.refreshToken()
	if err != nil {
		return err
	}
	wait := refreshDuration
	go func() {
		for {
			time.Sleep(wait)
			if err := interceptor.refreshToken(); err != nil {
				wait = time.Second
			} else {
				wait = refreshDuration
			}
		}
	}()
	return nil
}

func (interceptor *SimpleClientInterceptor) Unary() grpc.UnaryClientInterceptor {
	return interceptor.unaryClientInterceptor
}

func (interceptor *SimpleClientInterceptor) Stream() grpc.StreamClientInterceptor {
	return interceptor.streamClientInterceptor
}

func (interceptor *SimpleClientInterceptor) attachToken(ctx context.Context) context.Context {
	// read file saved token
	ctx = metadata.AppendToOutgoingContext(ctx, "authorization", interceptor.accessToken)
	return ctx
}

func (interceptor *SimpleClientInterceptor) unaryClientInterceptor(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic: %v", r)
		}
	}()

	start := time.Now()

	// attach jwt token
	log.Printf("interceptor: %+v\n", *interceptor)
	if interceptor.authMethods[method] {
		ctx = interceptor.attachToken(ctx)
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
	err = invoker(ctx, method, req, reply, cc, opts...)

	// get x-response-id header
	xrespid := header.Get("x-response-id")
	customHeader := header.Get("custom-resp-header")

	log.Printf("[gRPC client] Invoked RPC method=%s, xrid=%s, xrespid=%v, customHeader=%v, duration=%v, resp='%+v', error='%v'", method, xrid, xrespid, customHeader, time.Since(start), reply, err)

	return err
}

func (interceptor *SimpleClientInterceptor) streamClientInterceptor(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (clientStream grpc.ClientStream, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic: %v", r)
		}
	}()
	start := time.Now()

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

	var xrespid []string
	var customHeader []string
	header, ok := clientStream.Header()
	if ok == nil {
		xrespid = header.Get("x-response-id")
		customHeader = header.Get("custom-resp-header")
	}

	log.Printf("[gRPC client] Stream RPC method=%s, xrid=%s, xrespid=%v, customHeader=%v, duration=%v, error='%v'", method, xrid, xrespid, customHeader, time.Since(start), err)

	return clientStream, err
}

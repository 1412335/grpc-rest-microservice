package client

import (
	"context"
	"time"

	api_v3 "github.com/1412335/grpc-rest-microservice/pkg/api/v3"

	grpcClient "github.com/1412335/grpc-rest-microservice/pkg/client"
	"github.com/1412335/grpc-rest-microservice/pkg/configs"
	interceptor "github.com/1412335/grpc-rest-microservice/pkg/interceptor/client"
	"github.com/1412335/grpc-rest-microservice/pkg/log"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type User struct {
	ID   string `json:"id"`
	Role string `json:"role"`
}

type UserClient interface {
	Login(ctx context.Context, email, password string) (token string, err error)
	Validate(ctx context.Context, token string) (user *User, err error)
	Close() error
}

type userClientImpl struct {
	logger        log.Factory
	client        *grpcClient.Client
	userSrvClient api_v3.UserServiceClient
}

func NewUserServiceClient(cfgs *configs.ClientConfig, opt ...grpcClient.Option) (UserClient, error) {
	opt = append(opt,
		grpcClient.WithInterceptors(interceptor.NewSimpleClientInterceptor()),
	)
	client, err := grpcClient.New(cfgs, opt...)
	if err != nil {
		return nil, err
	}

	return &userClientImpl{
		logger:        client.GetLogger(),
		client:        client,
		userSrvClient: api_v3.NewUserServiceClient(client.ClientConn),
	}, nil
}

func (c *userClientImpl) Close() error {
	return c.client.Close()
}

func (c *userClientImpl) setHeader(ctx context.Context, m map[string]string) context.Context {
	md := metadata.New(m)
	return metadata.NewOutgoingContext(ctx, md)
}

func (c *userClientImpl) timeoutContext(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, 5*time.Second)
}

// login & get token
func (c *userClientImpl) Login(ctx context.Context, email, password string) (string, error) {
	ctx = c.setHeader(ctx, map[string]string{"custom-req-header": "login"})
	ctx, cancel := c.timeoutContext(ctx)
	defer cancel()
	// prepare request
	msg := &api_v3.LoginRequest{
		Email:    email,
		Password: password,
	}
	// fetch response headers
	var header metadata.MD
	// call service
	reply, err := c.userSrvClient.Login(ctx, msg, grpc.Header(&header))
	if err != nil {
		c.logger.For(ctx).Error("login failed", zap.Error(err))
		return "", err
	}
	// get response headers
	token := header.Get("token")
	if len(token) > 0 {
		c.logger.For(ctx).Info("response headers", zap.Strings("token", token))
	}
	// return
	return reply.GetToken(), nil
}

// validate token
func (c *userClientImpl) Validate(ctx context.Context, token string) (*User, error) {
	ctx, cancel := c.timeoutContext(ctx)
	defer cancel()
	// prepare request
	msg := &api_v3.ValidateRequest{
		Token: token,
	}
	// call service
	reply, err := c.userSrvClient.Validate(ctx, msg)
	if err != nil {
		c.logger.For(ctx).Error("validate token failed", zap.Error(err))
		return nil, err
	}
	return &User{
		ID:   reply.GetUser().GetId(),
		Role: reply.GetUser().GetRole().String(),
	}, nil
}

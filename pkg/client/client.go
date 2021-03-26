package client

import (
	"net"
	"strconv"

	"github.com/1412335/grpc-rest-microservice/pkg/configs"
	interceptor "github.com/1412335/grpc-rest-microservice/pkg/interceptor/client"
	"github.com/1412335/grpc-rest-microservice/pkg/log"
	"github.com/1412335/grpc-rest-microservice/pkg/tracing"
	otgrpc "github.com/opentracing-contrib/go-grpc"
	"github.com/uber/jaeger-lib/metrics"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type ClientOption func(*Client) error

func WithMetricsFactory(metricsFactory metrics.Factory) ClientOption {
	return func(c *Client) error {
		c.metricsFactory = metricsFactory
		return nil
	}
}

func WithLoggerFactory(logger log.Factory) ClientOption {
	return func(c *Client) error {
		c.logger = logger
		return nil
	}
}

func WithInterceptors(interceptor ...interceptor.ClientInterceptor) ClientOption {
	return func(c *Client) error {
		c.interceptors = append(c.interceptors, interceptor...)
		return nil
	}
}

type Client struct {
	config         *configs.ClientConfig
	ClientConn     *grpc.ClientConn
	logger         log.Factory
	metricsFactory metrics.Factory
	interceptors   []interceptor.ClientInterceptor
}

func New(cfgs *configs.ClientConfig, opt ...ClientOption) (*Client, error) {
	// create client
	client := &Client{
		config: cfgs,
		logger: log.DefaultLogger,
	}
	// set options
	for _, o := range opt {
		if err := o(client); err != nil {
			client.logger.Bg().Error("Set client option error", zap.Error(err))
			return nil, err
		}
	}

	// resolve grpc-server address
	addr := net.JoinHostPort(cfgs.GRPC.Host, strconv.Itoa(cfgs.GRPC.Port))

	// gRPC client options
	opts := []grpc.DialOption{
		grpc.WithInsecure(),
		// grpc.WithTransportCredentials(credentials.NewClientTLSFromCert(utils.CertPool, "")),
		// grpc.WithBlock(),
	}

	// add client interceptors
	opts = append(opts, client.buildInterceptors()...)

	callOptions := []grpc.CallOption{}
	if cfgs.GRPC.MaxCallRecvMsgSize > 0 {
		callOptions = append(callOptions, grpc.MaxCallRecvMsgSize(cfgs.GRPC.MaxCallRecvMsgSize))
	}
	if cfgs.GRPC.MaxCallSendMsgSize > 0 {
		callOptions = append(callOptions, grpc.MaxCallSendMsgSize(cfgs.GRPC.MaxCallSendMsgSize))
	}
	if len(callOptions) > 0 {
		opts = append(opts, grpc.WithDefaultCallOptions(callOptions...))
	}

	// connect grpc server
	conn, err := grpc.Dial(
		addr,
		opts...,
	)
	if err != nil {
		client.logger.Bg().Error("Dial grpc server failed", zap.Error(err))
		return nil, err
	}
	client.ClientConn = conn
	return client, nil
}

func (c *Client) buildInterceptors() []grpc.DialOption {
	var unaryInterceptors []grpc.UnaryClientInterceptor
	var streamInterceptors []grpc.StreamClientInterceptor
	if c.config.Tracing.Flag {
		// create tracer
		tracer := tracing.Init(c.config.ServiceName, c.metricsFactory, c.logger)
		// tracing interceptor
		unaryInterceptors = append(unaryInterceptors, otgrpc.OpenTracingClientInterceptor(tracer))
		streamInterceptors = append(streamInterceptors, otgrpc.OpenTracingStreamClientInterceptor(tracer))
	}

	// client interceptors
	for _, interceptor := range c.interceptors {
		unaryInterceptors = append(unaryInterceptors, interceptor.Unary())
		streamInterceptors = append(streamInterceptors, interceptor.Stream())
	}

	// create grpc server
	return []grpc.DialOption{
		grpc.WithChainUnaryInterceptor(unaryInterceptors...),
		grpc.WithChainStreamInterceptor(streamInterceptors...),
	}
}

func (c *Client) Close() error {
	return c.ClientConn.Close()
}

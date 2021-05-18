package client

import (
	"net"
	"strconv"

	"github.com/1412335/grpc-rest-microservice/pkg/configs"
	interceptor "github.com/1412335/grpc-rest-microservice/pkg/interceptor/client"
	"github.com/1412335/grpc-rest-microservice/pkg/log"
	"github.com/1412335/grpc-rest-microservice/pkg/tracing"
	"github.com/1412335/grpc-rest-microservice/pkg/utils"
	otgrpc "github.com/opentracing-contrib/go-grpc"
	"github.com/uber/jaeger-lib/metrics"
	"github.com/uber/jaeger-lib/metrics/expvar"
	"github.com/uber/jaeger-lib/metrics/prometheus"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type Option func(*Client) error

func WithLoggerFactory(logger log.Factory) Option {
	return func(c *Client) error {
		c.logger = logger
		return nil
	}
}

func WithInterceptors(interceptor ...interceptor.ClientInterceptor) Option {
	return func(c *Client) error {
		c.interceptors = append(c.interceptors, interceptor...)
		return nil
	}
}

type Client struct {
	config       *configs.ClientConfig
	ClientConn   *grpc.ClientConn
	logger       log.Factory
	interceptors []interceptor.ClientInterceptor
}

func New(cfgs *configs.ClientConfig, opt ...Option) (*Client, error) {
	// create client
	client := &Client{
		config: cfgs,
	}

	// set log w client service name + version
	client.setLogger()

	// set options
	if err := client.Init(opt...); err != nil {
		client.logger.Error("Set client option error", zap.Error(err))
		return nil, err
	}

	// resolve grpc-server address
	addr := net.JoinHostPort(cfgs.GRPC.Host, strconv.Itoa(cfgs.GRPC.Port))

	// gRPC client options
	opts := []grpc.DialOption{
		grpc.WithInsecure(),
	}

	// insecure
	if cfgs.EnableTLS && cfgs.TLSCert != nil {
		creds, err := client.loadClientTLSCredentials()
		if err != nil {
			client.logger.Error("Load client TLS credentials failed", zap.Error(err))
		} else {
			client.logger.Info("Load client TLS credentials")
			opts = []grpc.DialOption{
				grpc.WithTransportCredentials(creds),
				// grpc.WithBlock(),
			}
		}
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
		client.logger.Error("Dial grpc server failed", zap.Error(err))
		return nil, err
	}
	client.ClientConn = conn
	return client, nil
}

// override options
func (c *Client) Init(opt ...Option) error {
	for _, o := range opt {
		if err := o(c); err != nil {
			return err
		}
	}
	return nil
}

func (c *Client) loadClientTLSCredentials() (credentials.TransportCredentials, error) {
	config, err := utils.LoadClientTLSConfig(c.config.TLSCert.CACert)
	if err != nil {
		return nil, err
	}
	// Create the credentials and return it
	return credentials.NewTLS(config), nil
}

func (c *Client) tracingInterceptor() (grpc.UnaryClientInterceptor, grpc.StreamClientInterceptor) {
	// metrics
	var metricsFactory metrics.Factory
	if c.config.EnableTracing {
		if c.config.Tracing != nil && c.config.Tracing.Metrics == "expvar" {
			metricsFactory = expvar.NewFactory(10) // 10 buckets for histograms
			c.logger.Info("[Tracing] Using expvar as metrics backend")
		} else {
			metricsFactory = prometheus.New().Namespace(metrics.NSOptions{Name: "tracing", Tags: nil})
			c.logger.Info("[Tracing] Using prometheus as metrics backend")
		}
	}
	// create tracer
	tracer := tracing.Init(c.config.ServiceName, metricsFactory, c.logger)
	// tracing interceptor
	return otgrpc.OpenTracingClientInterceptor(tracer), otgrpc.OpenTracingStreamClientInterceptor(tracer)
}

func (c *Client) buildInterceptors() []grpc.DialOption {
	var unaryInterceptors []grpc.UnaryClientInterceptor
	var streamInterceptors []grpc.StreamClientInterceptor

	// tracing
	if c.config.EnableTracing {
		unaryTracing, streamTracing := c.tracingInterceptor()
		unaryInterceptors = append(unaryInterceptors, unaryTracing)
		streamInterceptors = append(streamInterceptors, streamTracing)
	}

	// client interceptors
	interceptor.DefaultLogger = c.logger.With(zap.String("interceptor-type", "client"))
	for _, i := range c.interceptors {
		unaryInterceptors = append(unaryInterceptors, i.Unary())
		streamInterceptors = append(streamInterceptors, i.Stream())
	}

	// create grpc server
	return []grpc.DialOption{
		grpc.WithChainUnaryInterceptor(unaryInterceptors...),
		grpc.WithChainStreamInterceptor(streamInterceptors...),
	}
}

func (c *Client) setLogger() {
	c.logger = log.DefaultLogger.With(zap.String("client-service", c.config.ServiceName), zap.String("client-version", c.config.Version))
}

func (c *Client) GetLogger() log.Factory {
	return c.logger
}

func (c *Client) Close() error {
	return c.ClientConn.Close()
}

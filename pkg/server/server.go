package server

import (
	"context"
	"crypto/tls"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/1412335/grpc-rest-microservice/pkg/configs"
	interceptor "github.com/1412335/grpc-rest-microservice/pkg/interceptor/server"
	"github.com/1412335/grpc-rest-microservice/pkg/log"
	"github.com/1412335/grpc-rest-microservice/pkg/tracing"
	otgrpc "github.com/opentracing-contrib/go-grpc"
	"github.com/uber/jaeger-lib/metrics"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"
)

type ServerOption func(*Server) error

func WithMetricsFactory(metricsFactory metrics.Factory) ServerOption {
	return func(s *Server) error {
		s.metricsFactory = metricsFactory
		return nil
	}
}

func WithLoggerFactory(logger log.Factory) ServerOption {
	return func(s *Server) error {
		s.logger = logger
		return nil
	}
}

func WithInterceptors(interceptor ...interceptor.ServerInterceptor) ServerOption {
	return func(s *Server) error {
		s.interceptors = append(s.interceptors, interceptor...)
		return nil
	}
}

type Server struct {
	config         *configs.ServiceConfig
	grpcServer     *grpc.Server
	logger         log.Factory
	metricsFactory metrics.Factory
	interceptors   []interceptor.ServerInterceptor
}

func NewServer(srvConfig *configs.ServiceConfig, opt ...ServerOption) *Server {
	// create server
	srv := &Server{
		config: srvConfig,
		logger: log.DefaultLogger,
	}
	// set options
	for _, o := range opt {
		if err := o(srv); err != nil {
			srv.logger.Bg().Fatal("Set server option error", zap.Error(err))
			return nil
		}
	}
	// server options
	opts := []grpc.ServerOption{}

	// insecure
	if srvConfig.Secure != nil && srvConfig.Secure.Flag {
		opts = append(opts, srv.insecureServer())
	}

	opts = append(opts, srv.buildServerInterceptors()...)
	// create grpc server
	srv.grpcServer = grpc.NewServer(opts...)
	// NOTE: gogo/protobuf is currently not working perfectly with server reflection
	// grpc reflection: use with evans
	reflection.Register(srv.grpcServer)
	return srv
}

func (s *Server) loadServerTLSCredentials() (credentials.TransportCredentials, error) {
	// Load server's certificate and private key
	serverCert, err := tls.LoadX509KeyPair(s.config.Secure.TLSCert.CertPem, s.config.Secure.TLSCert.KeyPem)
	if err != nil {
		return nil, err
	}

	// Create the credentials and return it
	config := &tls.Config{
		Certificates: []tls.Certificate{serverCert},
		ClientAuth:   tls.NoClientCert,
	}

	return credentials.NewTLS(config), nil
}

func (s *Server) insecureServer() grpc.ServerOption {
	creds, err := s.loadServerTLSCredentials()
	if err != nil {
		s.logger.Bg().Fatal("Failed to parse key pair:", zap.Error(err))
		return nil
	}
	// return grpc.Creds(credentials.NewServerTLSFromCert(&utils.Cert))
	return grpc.Creds(creds)
}

func (s *Server) buildServerInterceptors() []grpc.ServerOption {
	var unaryInterceptors []grpc.UnaryServerInterceptor
	var streamInterceptors []grpc.StreamServerInterceptor
	if s.config.Tracing.Flag {
		// create tracer
		tracer := tracing.Init(s.config.ServiceName, s.metricsFactory, s.logger)
		// tracing interceptor
		unaryInterceptors = append(unaryInterceptors, otgrpc.OpenTracingServerInterceptor(tracer))
		streamInterceptors = append(streamInterceptors, otgrpc.OpenTracingStreamServerInterceptor(tracer))
	}

	// server interceptor
	for _, interceptor := range s.interceptors {
		unaryInterceptors = append(unaryInterceptors, interceptor.Unary())
		streamInterceptors = append(streamInterceptors, interceptor.Stream())
	}

	// create grpc server
	return []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(unaryInterceptors...),
		grpc.ChainStreamInterceptor(streamInterceptors...),
	}
}

func (s *Server) Run(registerService func(*grpc.Server) error, stopper func()) error {
	ctx := context.Background()
	host := net.JoinHostPort("0.0.0.0", strconv.Itoa(s.config.GRPC.Port))
	// open tcp connect
	listen, err := net.Listen("tcp", host)
	if err != nil {
		s.logger.For(ctx).Error("Create tcp listener failed", zap.Error(err))
		return err
	}

	// register implementation services server
	if err := registerService(s.grpcServer); err != nil {
		s.logger.For(ctx).Error("Register service failed", zap.Error(err))
		return err
	}

	// graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		for sig := range c {
			s.logger.For(ctx).Error("Shutting down gRPC server", zap.Stringer("signal", sig))
			s.grpcServer.GracefulStop()
			stopper()
			<-ctx.Done()
		}
	}()

	s.logger.For(ctx).Info("Starting gRPC server", zap.String("at", host))

	// run grpc server
	return s.grpcServer.Serve(listen)
}

func (s *Server) Logger() log.Factory {
	return s.logger
}

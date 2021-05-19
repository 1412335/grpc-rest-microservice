package server

import (
	"context"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/1412335/grpc-rest-microservice/pkg/configs"
	interceptor "github.com/1412335/grpc-rest-microservice/pkg/interceptor/server"
	"github.com/1412335/grpc-rest-microservice/pkg/log"
	"github.com/1412335/grpc-rest-microservice/pkg/tracing"
	"github.com/1412335/grpc-rest-microservice/pkg/utils"
	"github.com/opentracing/opentracing-go"

	otgrpc "github.com/opentracing-contrib/go-grpc"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"
)

type Option func(*Server) error

func WithLoggerFactory(logger log.Factory) Option {
	return func(s *Server) error {
		s.logger = logger
		return nil
	}
}

func WithInterceptors(interceptor ...interceptor.ServerInterceptor) Option {
	return func(s *Server) error {
		s.interceptors = append(s.interceptors, interceptor...)
		return nil
	}
}

func WithTracer(tracer opentracing.Tracer) Option {
	return func(s *Server) error {
		s.tracer = tracer
		return nil
	}
}

type Server struct {
	config       *configs.ServiceConfig
	grpcServer   *grpc.Server
	logger       log.Factory
	tracer       opentracing.Tracer
	interceptors []interceptor.ServerInterceptor
}

func NewServer(srvConfig *configs.ServiceConfig, opt ...Option) *Server {
	// create server
	srv := &Server{
		config: srvConfig,
		tracer: tracing.DefaultTracer,
	}
	srv.setLogger()

	// set options
	if err := srv.Init(opt...); err != nil {
		srv.logger.Error("Init server error", zap.Error(err))
		return nil
	}

	// server options
	opts := []grpc.ServerOption{}

	// insecure
	if srvConfig.EnableTLS && srvConfig.TLSCert != nil {
		opts = append(opts, srv.insecureServer())
	}

	// interceptors
	opts = append(opts, srv.buildServerInterceptors()...)

	// create grpc server
	srv.grpcServer = grpc.NewServer(opts...)

	// NOTE: gogo/protobuf is currently not working perfectly with server reflection
	// grpc reflection: use with evans
	reflection.Register(srv.grpcServer)

	return srv
}

// override options
func (s *Server) Init(opt ...Option) error {
	for _, o := range opt {
		if err := o(s); err != nil {
			return err
		}
	}
	return nil
}

func (s *Server) setLogger() {
	s.logger = log.With(zap.String("service", s.config.ServiceName), zap.String("version", s.config.Version))
}

func (s *Server) loadServerTLSCredentials() (credentials.TransportCredentials, error) {
	config, err := utils.LoadServerTLSConfig(s.config.TLSCert.CertPem, s.config.TLSCert.KeyPem)
	if err != nil {
		return nil, err
	}
	// Create the credentials and return it
	return credentials.NewTLS(config), nil
}

func (s *Server) insecureServer() grpc.ServerOption {
	creds, err := s.loadServerTLSCredentials()
	if err != nil {
		s.logger.Fatal("Failed to parse key pair:", zap.Error(err))
		return nil
	}
	// return grpc.Creds(credentials.NewServerTLSFromCert(&utils.Cert))
	return grpc.Creds(creds)
}

func (s *Server) tracingInterceptor() (grpc.UnaryServerInterceptor, grpc.StreamServerInterceptor) {
	if tracing.DefaultTracer == nil {
		var metrics string
		if s.config.Tracing != nil {
			metrics = s.config.Tracing.Metrics
		}
		// create tracer
		tracing.DefaultTracer = tracing.Init(s.config.ServiceName, metrics, s.logger)
		s.tracer = tracing.DefaultTracer
	}
	// tracing interceptor
	return otgrpc.OpenTracingServerInterceptor(tracing.DefaultTracer), otgrpc.OpenTracingStreamServerInterceptor(tracing.DefaultTracer)
}

func (s *Server) buildServerInterceptors() []grpc.ServerOption {
	var unaryInterceptors []grpc.UnaryServerInterceptor
	var streamInterceptors []grpc.StreamServerInterceptor

	// tracing
	if s.config.EnableTracing {
		unaryTracing, streamTracing := s.tracingInterceptor()
		unaryInterceptors = append(unaryInterceptors, unaryTracing)
		streamInterceptors = append(streamInterceptors, streamTracing)
	}

	// server interceptor
	interceptor.DefaultLogger = s.logger.With(zap.String("interceptor-type", "server"))
	for _, i := range s.interceptors {
		unaryInterceptors = append(unaryInterceptors, i.Unary())
		streamInterceptors = append(streamInterceptors, i.Stream())
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
		select {
		case sig := <-c:
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

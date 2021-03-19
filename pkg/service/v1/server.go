package v1

import (
	"context"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	api_v1 "github.com/1412335/grpc-rest-microservice/pkg/api/v1"
	"github.com/1412335/grpc-rest-microservice/pkg/configs"
	"github.com/1412335/grpc-rest-microservice/pkg/dal/mysql"
	"github.com/1412335/grpc-rest-microservice/pkg/log"
	"github.com/1412335/grpc-rest-microservice/pkg/tracing"
	_ "github.com/go-sql-driver/mysql"
	otgrpc "github.com/opentracing-contrib/go-grpc"
	"github.com/uber/jaeger-lib/metrics"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type Server struct {
	config         *configs.ServiceConfig
	server         *grpc.Server
	logger         log.Factory
	metricsFactory metrics.Factory
}

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

func NewServer(srvConfig *configs.ServiceConfig, opt ...ServerOption) *Server {
	// create server
	srv := &Server{
		config: srvConfig,
	}
	// set options
	for _, o := range opt {
		if err := o(srv); err != nil {
			srv.logger.Bg().Fatal("Set server option error", zap.Error(err))
			return nil
		}
	}

	var unaryInterceptors []grpc.UnaryServerInterceptor
	var streamInterceptors []grpc.StreamServerInterceptor
	if srvConfig.Tracing.Flag {
		// create tracer
		tracer := tracing.Init(srvConfig.ServiceName, srv.metricsFactory, srv.logger)
		// tracing interceptor
		unaryInterceptors = append(unaryInterceptors, otgrpc.OpenTracingServerInterceptor(tracer))
		streamInterceptors = append(streamInterceptors, otgrpc.OpenTracingStreamServerInterceptor(tracer))
	}

	// create grpc server
	server := grpc.NewServer(
		grpc.ChainUnaryInterceptor(unaryInterceptors...),
		grpc.ChainStreamInterceptor(streamInterceptors...),
	)

	// grpc reflection: use with evans
	reflection.Register(server)

	srv.server = server
	return srv
}

func (s *Server) Run() error {
	ctx := context.Background()
	host := net.JoinHostPort("0.0.0.0", strconv.Itoa(s.config.GRPC.Port))
	// open tcp connect
	listen, err := net.Listen("tcp", host)
	if err != nil {
		s.logger.For(ctx).Error("Create tcp listener failed", zap.Error(err))
		return err
	}

	// connect mysql
	dal, err := mysql.NewDataAccessLayer(ctx, s.config.Database)
	if err != nil {
		s.logger.For(ctx).Fatal("Connect db failed", zap.Error(err))
		return err
	}
	defer dal.Disconnect()

	// implement service
	api := NewToDoServiceServer(s.config.Version, dal.GetDatabase(), s.logger)
	// register impl service
	api_v1.RegisterToDoServiceServer(s.server, api)

	// graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		for sig := range c {
			s.logger.For(ctx).Error("Shutting down gRPC server", zap.Stringer("signal", sig))
			s.server.GracefulStop()
			<-ctx.Done()
		}
	}()

	s.logger.For(ctx).Info("Starting gRPC server", zap.String("at", host))
	// run grpc server
	return s.server.Serve(listen)
}

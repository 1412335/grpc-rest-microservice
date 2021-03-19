package v1

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	api_v1 "github.com/1412335/grpc-rest-microservice/pkg/api/v1"
	"github.com/1412335/grpc-rest-microservice/pkg/configs"
	"github.com/1412335/grpc-rest-microservice/pkg/log"
	"github.com/1412335/grpc-rest-microservice/pkg/tracing"
	_ "github.com/go-sql-driver/mysql"
	otgrpc "github.com/opentracing-contrib/go-grpc"
	"github.com/uber/jaeger-lib/metrics"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type Server struct {
	host           string
	version        string
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
		host:    net.JoinHostPort("0.0.0.0", strconv.Itoa(srvConfig.GRPC.Port)),
		version: srvConfig.Version,
	}
	// set options
	for _, o := range opt {
		if err := o(srv); err != nil {
			srv.logger.Bg().Fatal("Set server option error", zap.Error(err))
			return nil
		}
	}

	// connect mysql
	param := "parseTime=true"
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?%s",
		srvConfig.Database.User,
		srvConfig.Database.Password,
		srvConfig.Database.Host,
		srvConfig.Database.Scheme,
		param,
	)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		srv.logger.Bg().Fatal("Connect db failed", zap.Error(err))
		return nil
	}
	defer db.Close()
	// https://github.com/go-sql-driver/mysql/
	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)

	var opts []grpc.ServerOption
	if srvConfig.Tracing.Flag {
		// create tracer
		tracer := tracing.Init(srvConfig.ServiceName, srv.metricsFactory, srv.logger)
		// grpc server opts with tracing interceptor
		opts = append(opts,
			grpc.UnaryInterceptor(otgrpc.OpenTracingServerInterceptor(tracer)),
			grpc.StreamInterceptor(otgrpc.OpenTracingStreamServerInterceptor(tracer)),
		)
	}
	// create grpc server
	server := grpc.NewServer(opts...)

	// implement service
	api := NewToDoServiceServer(db, srv.logger)
	// register impl service
	api_v1.RegisterToDoServiceServer(server, api)

	srv.server = server
	return srv
}

func (s *Server) Run() error {
	ctx := context.Background()
	// open tcp connect
	listen, err := net.Listen("tcp", s.host)
	if err != nil {
		s.logger.For(ctx).Error("Create tcp listener failed", zap.Error(err))
		return err
	}

	// graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		for sig := range c {
			s.logger.For(ctx).Error("Shutting down gRPC server", zap.Stringer("signal received", sig))
			s.server.GracefulStop()
			<-ctx.Done()
		}
	}()

	s.logger.For(ctx).Info("Starting gRPC server", zap.String("at", s.host))
	// run grpc server
	return s.server.Serve(listen)
}

package v3

import (
	"context"

	api_v3 "github.com/1412335/grpc-rest-microservice/pkg/api/v3"
	"github.com/1412335/grpc-rest-microservice/pkg/cache"
	"github.com/1412335/grpc-rest-microservice/pkg/configs"
	"github.com/1412335/grpc-rest-microservice/pkg/dal/postgres"
	"github.com/1412335/grpc-rest-microservice/pkg/log"
	"github.com/1412335/grpc-rest-microservice/pkg/server"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

var (
	DefaultCache cache.Cache
)

type Server struct {
	server   *server.Server
	tokenSrv *TokenService
	dal      *postgres.DataAccessLayer
}

func NewServer(srvConfig *configs.ServiceConfig, logger log.Factory, opt ...server.ServerOption) *Server {
	// simple server interceptor
	// simpleInterceptor := &interceptor.SimpleServerInterceptor{}
	// opt = append(opt, server.WithInterceptors(simpleInterceptor))
	// simpleInterceptor.WithLogger(srv.logger)

	// init postgres
	dal, err := postgres.NewDataAccessLayer(context.Background(), srvConfig.Database)
	if err != nil || dal.GetDatabase() == nil {
		logger.Bg().Error("init db failed", zap.Error(err))
		return nil
	}
	// migrate model
	if err := dal.GetDatabase().AutoMigrate(&User{}); err != nil {
		logger.Bg().Error("migrate db failed", zap.Error(err))
		return nil
	}

	// create server
	srv := &Server{
		tokenSrv: NewTokenService(srvConfig.JWT),
		dal:      dal,
	}

	// auth server interceptor
	authInterceptor := NewAuthServerInterceptor(logger, srv.tokenSrv, srvConfig.AccessibleRoles)

	// append server options with logger + auth token interceptor
	opt = append(opt,
		server.WithInterceptors(authInterceptor),
		server.WithLoggerFactory(logger),
	)

	// grpc server
	s := server.NewServer(srvConfig, opt...)

	srv.server = s
	return srv
}

func (s *Server) Run() error {
	return s.server.Run(func(srv *grpc.Server) error {
		// implement service
		api := NewUserService(s.dal, s.server.Logger(), s.tokenSrv)

		// register impl service
		api_v3.RegisterUserServiceServer(srv, api)
		return nil
	}, func() {
		// close db connection
		defer s.dal.Disconnect()
	})
}

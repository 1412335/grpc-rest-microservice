package v3

import (
	"context"

	api_v3 "github.com/1412335/grpc-rest-microservice/pkg/api/v3"
	"github.com/1412335/grpc-rest-microservice/pkg/configs"
	"github.com/1412335/grpc-rest-microservice/pkg/dal/postgres"
	"github.com/1412335/grpc-rest-microservice/pkg/server"
	"google.golang.org/grpc"
)

type Server struct {
	server   *server.Server
	tokenSrv *TokenService
	dal      *postgres.DataAccessLayer
}

func NewServer(srvConfig *configs.ServiceConfig, opt ...server.ServerOption) *Server {
	// simple server interceptor
	// simpleInterceptor := &interceptor.SimpleServerInterceptor{}
	// opt = append(opt, server.WithInterceptors(simpleInterceptor))
	// simpleInterceptor.WithLogger(srv.logger)

	// dal
	dal, err := postgres.NewDataAccessLayer(context.Background(), srvConfig.Database)
	if err != nil {
		return nil
	}

	// create server
	srv := &Server{
		tokenSrv: NewTokenService(srvConfig.JWT),
		dal:      dal,
	}

	// // auth server interceptor
	// authInterceptor := NewAuthServerInterceptor(s.server.logger, srv.tokenSrv, srvConfig.AccessibleRoles)
	// opt = append(opt, server.WithInterceptors(authInterceptor))

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
		defer s.dal.Disconnect()
	})
}

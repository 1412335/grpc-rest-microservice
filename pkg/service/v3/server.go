package v3

import (
	api_v3 "github.com/1412335/grpc-rest-microservice/pkg/api/v3"
	"github.com/1412335/grpc-rest-microservice/pkg/configs"
	"github.com/1412335/grpc-rest-microservice/pkg/dal/mysql"
	"github.com/1412335/grpc-rest-microservice/pkg/interceptor"
	"github.com/1412335/grpc-rest-microservice/pkg/server"
	"github.com/1412335/grpc-rest-microservice/pkg/utils"
	_ "github.com/go-sql-driver/mysql"
	"google.golang.org/grpc"
)

type Server struct {
	server     *server.Server
	jwtManager *utils.JWTManager
	dal        *mysql.DataAccessLayer
}

func NewServer(srvConfig *configs.ServiceConfig, opt ...server.ServerOption) *Server {
	// simple server interceptor
	simpleInterceptor := &interceptor.SimpleServerInterceptor{}
	opt = append(opt, server.WithInterceptors(simpleInterceptor))
	// simpleInterceptor.WithLogger(srv.logger)

	// auth server interceptor
	// authInterceptor := interceptor.NewAuthServerInterceptor(srv.logger, srv.jwtManager, srvConfig.AccessibleRoles)
	// opt = append(opt, server.authInterceptor(simpleInterceptor))

	// create server
	s := server.NewServer(srvConfig, opt...)
	srv := &Server{
		server:     s,
		jwtManager: utils.NewJWTManager(srvConfig.JWT),
	}
	return srv
}

func (s *Server) Run() error {
	return s.server.Run(func(srv *grpc.Server) error {
		// implement service
		api := NewUserService(s.server.Logger())

		// register impl service
		api_v3.RegisterUserServiceServer(srv, api)
		return nil
	})
}

package account

import (
	"context"

	pb "account/api"
	"account/model"

	"github.com/1412335/grpc-rest-microservice/pkg/configs"
	"github.com/1412335/grpc-rest-microservice/pkg/dal/postgres"
	"github.com/1412335/grpc-rest-microservice/pkg/dal/redis"
	"github.com/1412335/grpc-rest-microservice/pkg/log"
	"github.com/1412335/grpc-rest-microservice/pkg/server"
	"github.com/1412335/grpc-rest-microservice/v3/pkg/cache"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type Server struct {
	server *server.Server
	dal    *postgres.DataAccessLayer
}

func NewServer(srvConfig *configs.ServiceConfig, opt ...server.Option) *Server {
	// init postgres
	dal, err := postgres.NewDataAccessLayer(context.Background(), srvConfig.Database)
	if err != nil || dal.GetDatabase() == nil {
		log.Error("init db failed", zap.Error(err))
		return nil
	}
	// migrate db
	if err := dal.GetDatabase().AutoMigrate(
		&model.Account{},
	); err != nil {
		log.Error("migrate db failed", zap.Error(err))
		return nil
	}

	// connect redis
	redisStore, err := redis.New(redis.WithNodes(srvConfig.Redis.Nodes), redis.WithPrefix(srvConfig.ServiceName))
	if err != nil {
		log.Error("connect redis store failed", zap.Error(err))
	} else if redisStore != nil {
		// cache w redis store
		cache.DefaultCache, err = cache.NewRedisCache(redisStore, cache.WithPrefix(srvConfig.ServiceName))
		if err != nil {
			log.Error("create cache redis store failed", zap.Error(err))
		}
	}

	// create server
	srv := &Server{
		dal: dal,
	}

	// // auth server interceptor
	// authInterceptor := NewAuthServerInterceptor(srv.tokenSrv, srvConfig.AuthRequiredMethods, srvConfig.AccessibleRoles)

	// // append server options with logger + auth token interceptor
	// opt = append(opt,
	// 	server.WithInterceptors(
	// 		// interceptor.NewSimpleServerInterceptor(),
	// 		authInterceptor,
	// 	),
	// )

	// grpc server
	s := server.NewServer(srvConfig, opt...)

	srv.server = s
	return srv
}

func (s *Server) Run() error {
	return s.server.Run(func(srv *grpc.Server) error {
		log.Info("Register", zap.String("service", "account"))

		// implement service
		api := NewAccountService(s.dal)

		// register impl service
		pb.RegisterAccountServiceServer(srv, api)
		return nil
	}, func() {
		// close db connection
		defer s.dal.Disconnect()
	})
}

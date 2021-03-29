package cmd

import (
	"time"

	"github.com/1412335/grpc-rest-microservice/pkg/cache"
	grpcClient "github.com/1412335/grpc-rest-microservice/pkg/client"
	"github.com/1412335/grpc-rest-microservice/pkg/configs"
	"github.com/1412335/grpc-rest-microservice/pkg/dal/redis"
	"github.com/1412335/grpc-rest-microservice/pkg/log"
	"github.com/1412335/grpc-rest-microservice/pkg/server"
	v3 "github.com/1412335/grpc-rest-microservice/service/v3"
	"github.com/1412335/grpc-rest-microservice/service/v3/client"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var v3Cmd = &cobra.Command{
	Use:   "v3",
	Short: "Start Service version 3",
	Long:  `Start Service version 3`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return V3Service()
	},
}

func init() {
	logger.Info("v3.Init")
	rootCmd.AddCommand(v3Cmd)
}

func V3Service() error {
	// create log factory
	zapLogger := logger.With(zap.String("service", cfgs.ServiceName), zap.String("version", cfgs.Version))
	logger := log.NewFactory(zapLogger)

	// cache
	if redisStore, err := redis.New(redis.WithNodes(cfgs.Redis.Nodes), redis.WithPrefix(cfgs.ServiceName)); err != nil {
		zapLogger.Error("Connect redis store failed", zap.Error(err))
	} else if cache, err := cache.NewRedisCache(redisStore, cache.WithExpiryDuration(120*time.Second)); err != nil {
		zapLogger.Error("Create cache w redis store failed", zap.Error(err))
	} else {
		v3.DefaultCache = cache
	}

	// server
	server := v3.NewServer(
		cfgs,
		logger,
		server.WithMetricsFactory(metricsFactory),
	)

	// run grpc server
	// return logError(zapLogger, server.Run())
	go func() {
		logError(zapLogger, server.Run())
	}()

	// go func() {
	// 	err := testGrpcClient(cfgs.ClientConfig, logger)
	// 	if err != nil {
	// 		logError(zapLogger, err)
	// 	}
	// }()

	// run grpc-gateway
	handler := v3.NewHandler(cfgs)
	err := handler.Run()
	if err != nil {
		zapLogger.Error("Starting gRPC-gateway error", zap.Error(err))
	}
	return err
}

func testGrpcClient(cfgs *configs.ClientConfig, logger log.Factory) error {
	var opts []grpcClient.ClientOption
	if cfgs.EnableTracing {
		opts = append(opts, grpcClient.WithMetricsFactory(metricsFactory))
	}
	c, err := client.New(
		cfgs,
		logger,
		opts...,
	)
	if err != nil {
		return err
	}
	defer c.Close()

	// login
	username, password := "string@gmail.com", "stringstring"
	if token, err := c.Login(username, password); err != nil {
		return err
	} else {
		logger.Bg().Info("login resp", zap.String("token", token))
	}
	return nil
}

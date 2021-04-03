package cmd

import (
	"errors"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	grpcClient "github.com/1412335/grpc-rest-microservice/pkg/client"
	"github.com/1412335/grpc-rest-microservice/pkg/log"
	"github.com/1412335/grpc-rest-microservice/service/v3/client"
)

var v3clientCmd = &cobra.Command{
	Use:   "v3-client",
	Short: "Start grpc client for api v3",
	Long:  `Start grpc client for api v3`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return V3ClientService()
	},
}

func init() {
	// logger.Info("client.Init")
	rootCmd.AddCommand(v3clientCmd)
}

func V3ClientService() error {
	// create log factory
	logger := log.DefaultLogger.With(
		zap.String("service", cfgs.ServiceName),
		zap.String("version", cfgs.Version),
	)
	// get user client configs
	clientCfgs, ok := cfgs.ClientConfig["user"]
	if !ok {
		return logError(logger, errors.New("not found user client config"))
	}
	zapLogger := logger.With(
		zap.String("client-service", clientCfgs.ServiceName),
		zap.String("client-service-version", clientCfgs.Version),
	)

	// set default logger
	// v3.DefaultLogger = zapLogger

	var opts []grpcClient.Option
	c, err := client.New(
		clientCfgs,
		opts...,
	)

	if err != nil {
		return logError(zapLogger, err)
	}
	defer c.Close()

	// login
	username, password := "string@gmail.com", "stringstring"
	if token, err := c.Login(username, password); err != nil {
		return logError(zapLogger, err)
	} else {
		zapLogger.Bg().Info("login resp", zap.String("token", token))
	}

	return nil
}

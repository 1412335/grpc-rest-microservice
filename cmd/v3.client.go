package cmd

import (
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	grpcClient "github.com/1412335/grpc-rest-microservice/pkg/client"
	"github.com/1412335/grpc-rest-microservice/pkg/configs"
	"github.com/1412335/grpc-rest-microservice/pkg/log"
	"github.com/1412335/grpc-rest-microservice/pkg/service/v3/client"
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
	zapLogger := logger.With(
		zap.String("service", cfgs.ServiceName),
		zap.String("version", cfgs.Version),
		zap.String("client-service", cfgs.ClientConfig.ServiceName),
		zap.String("client-service-version", cfgs.ClientConfig.Version),
	)
	logger := log.NewFactory(zapLogger)

	// inherit tracing flag
	if cfgs.ClientConfig.Tracing == nil {
		cfgs.ClientConfig.Tracing = &configs.Tracing{
			Flag: cfgs.Tracing.Flag,
		}
	}

	c, err := client.New(
		cfgs.ClientConfig,
		logger,
		grpcClient.WithMetricsFactory(metricsFactory),
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
		zapLogger.Info("login resp", zap.String("token", token))
	}

	return nil
}

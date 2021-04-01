package cmd

import (
	"github.com/1412335/grpc-rest-microservice/pkg/log"
	v1 "github.com/1412335/grpc-rest-microservice/service/v1"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var v1Cmd = &cobra.Command{
	Use:   "v1",
	Short: "Start Service version 1",
	Long:  `Start Service version 1`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return V1Service()
	},
}

func init() {
	logger.Info("v1.Init")
	rootCmd.AddCommand(v1Cmd)
}

func V1Service() error {
	// create log factory
	logger := log.DefaultLogger.With(zap.String("service", cfgs.ServiceName), zap.String("version", cfgs.Version))
	// server
	server := v1.NewServer(
		cfgs,
		v1.WithMetricsFactory(metricsFactory),
		v1.WithLoggerFactory(logger),
	)
	return logError(logger, server.Run())
}

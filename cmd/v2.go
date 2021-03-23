package cmd

import (
	"github.com/1412335/grpc-rest-microservice/pkg/log"
	v2 "github.com/1412335/grpc-rest-microservice/pkg/service/v2"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var v2Cmd = &cobra.Command{
	Use:   "v2",
	Short: "Start Service version 2",
	Long:  `Start Service version 2`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return V2Service()
	},
}

func init() {
	logger.Info("v2.Init")
	rootCmd.AddCommand(v2Cmd)
}

func V2Service() error {
	// create log factory
	zapLogger := logger.With(zap.String("service", cfgs.ServiceName), zap.String("version", cfgs.Version))
	logger := log.NewFactory(zapLogger)
	// server
	server := v2.NewServer(
		cfgs,
		v2.WithMetricsFactory(metricsFactory),
		v2.WithLoggerFactory(logger),
	)

	// run grpc server
	// return logError(zapLogger, server.Run())
	go func() {
		logError(zapLogger, server.Run())
	}()

	// run grpc-gateway
	handler := v2.NewHandler(cfgs)
	err := handler.Run()
	if err != nil {
		zapLogger.Error("Starting gRPC-gateway error", zap.Error(err))
	}
	return err
}
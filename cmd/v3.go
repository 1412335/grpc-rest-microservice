package cmd

import (
	"github.com/1412335/grpc-rest-microservice/pkg/log"
	"github.com/1412335/grpc-rest-microservice/pkg/server"
	v3 "github.com/1412335/grpc-rest-microservice/pkg/service/v3"

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

	// run grpc-gateway
	handler := v3.NewHandler(cfgs)
	err := handler.Run()
	if err != nil {
		zapLogger.Error("Starting gRPC-gateway error", zap.Error(err))
	}
	return err
}

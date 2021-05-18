package cmd

import (
	"github.com/1412335/grpc-rest-microservice/pkg/log"
	v2 "github.com/1412335/grpc-rest-microservice/service/v2"

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
	log.Info("v2.Init")
	rootCmd.AddCommand(v2Cmd)
}

func V2Service() error {
	// create log factory
	logger := log.With(zap.String("service", cfgs.ServiceName), zap.String("version", cfgs.Version))
	// server
	server := v2.NewServer(
		cfgs,
		v2.WithLoggerFactory(logger),
	)

	// run grpc server
	// return logError(zapLogger, server.Run())
	go func() {
		if err := server.Run(); err != nil {
			logger.Bg().Error("Server running failed", zap.Error(err))
		}
	}()

	// run grpc-gateway
	handler := v2.NewHandler(cfgs)
	err := handler.Run()
	if err != nil {
		logger.Bg().Error("Starting gRPC-gateway error", zap.Error(err))
	}
	return err
}

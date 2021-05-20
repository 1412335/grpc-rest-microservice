package cmd

import (
	grpcClient "github.com/1412335/grpc-rest-microservice/pkg/client"
	"github.com/1412335/grpc-rest-microservice/pkg/configs"
	"github.com/1412335/grpc-rest-microservice/pkg/log"
	v3 "github.com/1412335/grpc-rest-microservice/service/v3"
	"github.com/1412335/grpc-rest-microservice/service/v3/client"
	"github.com/1412335/grpc-rest-microservice/service/v3/handler"

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
	log.Info("v3.Init")
	rootCmd.AddCommand(v3Cmd)
}

func V3Service() error {
	// create log factory
	zapLogger := log.With(zap.String("service", cfgs.ServiceName), zap.String("version", cfgs.Version))
	// set default logger
	// v3.DefaultLogger = zapLogger

	// server
	server := v3.NewServer(
		cfgs,
	)

	// run grpc server
	// return logError(zapLogger, server.Run())
	go func() {
		if err := server.Run(); err != nil {
			zapLogger.Bg().Error("Server running failed", zap.Error(err))
		}
	}()

	// go func() {
	// 	err := testGrpcClient(cfgs.ClientConfig["user"])
	// 	if err != nil {
	// 		logError(zapLogger, err)
	// 	}
	// }()

	// run grpc-gateway
	handler := handler.NewHandler(cfgs)
	err := handler.Run()
	if err != nil {
		zapLogger.Error("Starting gRPC-gateway error", zap.Error(err))
	}
	return err
}

//nolint:unused
func testGrpcClient(cfgs *configs.ClientConfig) error {
	var opts []grpcClient.Option
	c, err := client.New(
		cfgs,
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
		log.Info("login resp", zap.String("token", token))
	}
	return nil
}

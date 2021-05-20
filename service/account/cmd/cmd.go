package cmd

import (
	handlerSrv "account/handler"
	serverSrv "account/server"

	"github.com/1412335/grpc-rest-microservice/cmd"
	"github.com/1412335/grpc-rest-microservice/pkg/log"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var command = &cobra.Command{
	Use:   "start-service",
	Short: "Start Service",
	Long:  `Start Service`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return RunService()
	},
}

func init() {
	log.Info("account.Init")
	cmd.AddCommand(command)
}

func Execute() {
	cmd.Execute()
}

func RunService() error {
	// load service configs
	cfgs := cmd.LoadConfig()

	// create log factory
	zapLogger := log.With(zap.String("service", cfgs.ServiceName), zap.String("version", cfgs.Version))
	// set default logger
	// log.DefaultLogger = zapLogger

	// run grpc server
	go func() {
		// server
		server := serverSrv.NewServer(
			cfgs,
		)
		if err := server.Run(); err != nil {
			zapLogger.Bg().Error("server run failed", zap.Error(err))
		}
	}()

	// go func() {
	// 	err := (func(cfgs *configs.ClientConfig) error {
	// 		var opts []grpcClient.Option
	// 		c, err := client.New(
	// 			cfgs,
	// 			opts...,
	// 		)
	// 		if err != nil {
	// 			return err
	// 		}
	// 		defer c.Close()

	// 		// login
	// 		username, password := "string@gmail.com", "stringstring"
	// 		if token, err := c.Login(username, password); err != nil {
	// 			return err
	// 		} else {
	// 			log.Info("login resp", zap.String("token", token))
	// 		}
	// 		return nil
	// 	})(cfgs.ClientConfig["user"])
	// 	if err != nil {
	// 		logError(zapLogger, err)
	// 	}
	// }()

	// run grpc-gateway
	handler := handlerSrv.NewHandler(cfgs)
	err := handler.Run()
	if err != nil {
		zapLogger.Error("Starting gRPC-gateway error", zap.Error(err))
	}
	return err
}

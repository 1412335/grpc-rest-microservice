package cmd

import (
	"errors"
	"net"
	"strconv"

	"time"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	api_v2 "github.com/1412335/grpc-rest-microservice/pkg/api/v2/grpc-gateway/gen"
	"github.com/1412335/grpc-rest-microservice/pkg/bridge"
	"github.com/1412335/grpc-rest-microservice/pkg/log"
)

var clientCmd = &cobra.Command{
	Use:   "client",
	Short: "Start grpc client for api v2",
	Long:  `Start grpc client for api v2`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return ClientService()
	},
}

func init() {
	// logger.Info("client.Init")
	rootCmd.AddCommand(clientCmd)
}

func ClientService() error {
	// create log factory
	zapLogger := log.DefaultLogger.With(zap.String("client-service", cfgs.ServiceName), zap.String("version", cfgs.Version))
	// logger := log.NewFactory(zapLogger)

	//
	managerClient := bridge.NewManagerClientWithConfigs(cfgs.ManagerClient)
	if managerClient == nil {
		return logError(zapLogger, errors.New("Create new manager client failed"))
	}

	addr := net.JoinHostPort(cfgs.GRPC.Host, strconv.Itoa(cfgs.GRPC.Port))
	client, err := managerClient.GetClient(addr)
	if err != nil {
		return logError(zapLogger, err)
	}

	// unary request
	timestamp := int64(2222)
	if resp, err := client.Ping(timestamp); err != nil {
		zapLogger.Bg().Error("Ping error", zap.Error(err))
		return err
	} else {
		zapLogger.Bg().Info("Ping resp", zap.Any("resp", resp))
	}

	for i := int64(0); i < 2; i++ {
		if resp, err := client.Post(timestamp + i); err != nil {
			zapLogger.Bg().Error("Post error", zap.Error(err))
			return err
		} else {
			zapLogger.Bg().Info("Post resp", zap.Any("resp", resp))
		}
		time.Sleep(time.Second)
	}

	// server streaming
	count := int32(2)
	interval := int32(100)
	if resp, err := client.StreamingPing(timestamp, count, interval); err != nil {
		zapLogger.Bg().Error("Server streaming ping error", zap.Error(err))
		return err
	} else {
		zapLogger.Bg().Info("Server streaming ping resp", zap.Any("resp", resp))
	}

	// client streaming
	msg := []*api_v2.StreamingMessagePing{
		{
			Timestamp:    1,
			MessageCount: 1,
		},
		{
			Timestamp:    2,
			MessageCount: 2,
		},
	}
	if resp, err := client.StreamingPost(msg); err != nil {
		zapLogger.Bg().Error("Client streaming post error", zap.Error(err))
		return err
	} else {
		zapLogger.Bg().Info("Client streaming post resp", zap.Any("resp", resp))
	}

	// duplex streaming
	if resp, err := client.DuplexStreaming(msg); err != nil {
		zapLogger.Bg().Error("Duplex streaming post error", zap.Error(err))
		return err
	} else {
		zapLogger.Bg().Info("Duplex streaming post resp", zap.Any("resp", resp))
	}

	return nil
}

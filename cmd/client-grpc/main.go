package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/1412335/grpc-rest-microservice/pkg/api/v2/bridge"
	api_v2 "github.com/1412335/grpc-rest-microservice/pkg/api/v2/gen/grpc-gateway/gen"
	"github.com/1412335/grpc-rest-microservice/pkg/configs"
)

var (
	address = flag.String("grpc-host", "", "gRPC server in format host:port")
)

// using pool connection with grpcpool
func poolConnections() {
	// load config using viper
	clientConfigs := &configs.ServiceConfig{}
	if err := configs.LoadConfig("", clientConfigs); err != nil {
		log.Fatalf("[Main] Load config failed: %v", err)
	}
	// if err := viper.Unmarshal(clientConfigs); err != nil {
	// 	log.Fatalf("[Main] Unmarshal config failed: %v", err)
	// }

	managerClient := bridge.NewManagerClientWithConfigs(clientConfigs.ManagerClient)
	if managerClient == nil {
		log.Fatalf("[Main] Create new manager client failed")
	}

	*address = fmt.Sprintf("%s:%d", clientConfigs.GRPC.Host, clientConfigs.GRPC.Port)
	client, err := managerClient.GetClient(*address)
	if err != nil {
		log.Fatalf("[Main] Get client error: %+v", err)
	}

	// unary request
	timestamp := int64(2222)
	if resp, err := client.Ping(timestamp); err != nil {
		log.Fatalf("[Test] Ping error: %+v", err)
	} else {
		log.Printf("[Test] Ping resp: %+v\n", resp)
	}

	for i := int64(0); i < 2; i++ {
		if resp, err := client.Post(timestamp + i); err != nil {
			log.Fatalf("[Test] Post error: %+v", err)
		} else {
			log.Printf("[Test] Post resp: %+v\n", resp)
		}
		time.Sleep(time.Second)
	}

	// server streaming
	count := int32(2)
	interval := int32(100)
	if resp, err := client.StreamingPing(timestamp, count, interval); err != nil {
		log.Fatalf("[Test] Server Streaming Ping error: %+v", err)
	} else {
		log.Printf("[Test] Server Streaming Ping resp: %+v\n", resp)
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
		log.Fatalf("[Test] Client Streaming Post error: %+v", err)
	} else {
		log.Printf("[Test] Client Streaming Post resp: %+v\n", resp)
	}

	// duplex streaming
	if resp, err := client.DuplexStreaming(msg); err != nil {
		log.Fatalf("[Test] Duplex Streaming error: %+v", err)
	} else {
		log.Printf("[Test] Duplex Streaming resp: %+v\n", resp)
	}
}

func main() {
	flag.Parse()
	poolConnections()
}

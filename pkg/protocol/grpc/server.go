package grpc

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	api_v1 "grpc-rest-microservice/pkg/api/v1"
	api_v2 "grpc-rest-microservice/pkg/api/v2/gen/grpc-gateway/gen"

	"google.golang.org/grpc"
)

func RunServer(ctx context.Context, v1API api_v1.ToDoServiceServer, port string) error {
	listen, err := net.Listen("tcp", fmt.Sprintf(":%v", port))
	if err != nil {
		return err
	}

	// register implementation service
	server := grpc.NewServer()
	api_v1.RegisterToDoServiceServer(server, v1API)

	// graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGKILL)
	go func() {
		for {
			select {
			case <-c:
				log.Println("Shutting down gRPC server...")
				server.GracefulStop()

				<-ctx.Done()
			default:
			}
		}
	}()

	log.Println("starting gRPC server...")
	return server.Serve(listen)
}

func RunServerV2(ctx context.Context, v2API api_v2.ServiceAServer, v2API_extra api_v2.ServiceExtraServer, port string) error {
	listen, err := net.Listen("tcp", fmt.Sprintf(":%v", port))
	if err != nil {
		return err
	}

	// register implementation service
	server := grpc.NewServer()

	api_v2.RegisterServiceAServer(server, v2API)
	api_v2.RegisterServiceExtraServer(server, v2API_extra)

	// graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGKILL)
	go func() {
		for {
			select {
			case <-c:
				log.Println("Shutting down gRPC server v2...")
				server.GracefulStop()

				<-ctx.Done()
			default:
			}
		}
	}()

	log.Println("starting gRPC server v2 at:", port)
	return server.Serve(listen)
}

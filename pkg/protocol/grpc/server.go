package grpc

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	api_v1 "github.com/1412335/grpc-rest-microservice/pkg/api/v1"
	api_v2 "github.com/1412335/grpc-rest-microservice/pkg/api/v2/gen/grpc-gateway/gen"
	"github.com/1412335/grpc-rest-microservice/pkg/interceptor"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
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
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		for range c {
			log.Println("Shutting down gRPC server...")
			server.GracefulStop()

			<-ctx.Done()
		}
	}()

	log.Println("starting gRPC server...")
	return server.Serve(listen)
}

func RunServerV2(ctx context.Context, serverInterceptor interceptor.ServerInterceptor, v2API api_v2.ServiceAServer, v2APIExtra api_v2.ServiceExtraServer, port string) error {
	listen, err := net.Listen("tcp", fmt.Sprintf(":%v", port))
	if err != nil {
		return err
	}

	// register implementation service with interceptor
	server := grpc.NewServer(
		grpc.UnaryInterceptor(serverInterceptor.Unary()),
		grpc.StreamInterceptor(serverInterceptor.Stream()),
	)

	api_v2.RegisterServiceAServer(server, v2API)
	api_v2.RegisterServiceExtraServer(server, v2APIExtra)

	// grpc reflection
	reflection.Register(server)

	// graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		for range c {
			log.Println("Shutting down gRPC server v2...")
			server.GracefulStop()

			<-ctx.Done()
		}
	}()

	log.Println("starting gRPC server v2 at:", port)
	return server.Serve(listen)
}

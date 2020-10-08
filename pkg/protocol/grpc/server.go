package grpc

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"

	api_v1 "grpc-rest-microservice/pkg/api/v1"
	api_v2 "grpc-rest-microservice/pkg/api/v2/gen/grpc-gateway/gen"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
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

// middleware for unary request to grpc server
func interceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {

	var xrid []string
	var customHeader []string
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic: %v", r)
		}
	}()
	defer func() {
		log.Printf("[gRPC server] Received RPC method=%s, xrid=%v, customHeader=%v, error='%v'", info.FullMethod, xrid, customHeader, err)
	}()

	// fetch headers req
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.DataLoss, "failed to get metadata")
	}
	xrid = md.Get("x-request-id")
	if len(xrid) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "missing 'x-request-id' header")
	}
	if strings.Trim(xrid[0], " ") == "" {
		return nil, status.Errorf(codes.InvalidArgument, "empty 'x-request-id' header")
	}

	// fetch custom-request-header
	customHeader = md.Get("custom-req-header")

	// validate request
	// log.Println("[gRPC server] validate req")

	// send x-response-id header
	header := metadata.New(map[string]string{"x-response-id": xrid[0]})
	if err := grpc.SendHeader(ctx, header); err != nil {
		return nil, status.Errorf(codes.Internal, "unable to send 'x-response-id' header")
	}
	// NOT WORK: because server service does NOT using context to send anything
	// ctx = metadata.AppendToOutgoingContext(ctx, []string{"x-response-id", xrid[0]}...)

	resp, err = handler(ctx, req)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "serve handler error: %+v", err)
	}

	// add serviceName into response
	if msg, ok := resp.(*api_v2.MessagePong); ok {
		msg.ServiceName = info.FullMethod
		return msg, nil
	}

	return resp, nil
}

func RunServerV2(ctx context.Context, v2API api_v2.ServiceAServer, v2API_extra api_v2.ServiceExtraServer, port string) error {
	listen, err := net.Listen("tcp", fmt.Sprintf(":%v", port))
	if err != nil {
		return err
	}

	// register implementation service with interceptor
	server := grpc.NewServer(grpc.UnaryInterceptor(interceptor))

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

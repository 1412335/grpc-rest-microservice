package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"strings"
	"sync"
	"time"

	api_v1 "grpc-rest-microservice/pkg/api/v1"
	"grpc-rest-microservice/pkg/api/v2/bridge"
	api_v2 "grpc-rest-microservice/pkg/api/v2/gen/grpc-gateway/gen"

	"github.com/golang/protobuf/ptypes"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const (
	apiVersion = "v1"
)

func testingAPI_V1(ctx context.Context, client api_v1.ToDoServiceClient) {
	log.Println("starting test api_v1....")

	t := time.Now().In(time.UTC)
	reminder, _ := ptypes.TimestampProto(t)
	pfx := t.Format(time.RFC3339Nano)

	// create todo
	req1 := api_v1.CreateRequest{
		Api: apiVersion,
		ToDo: &api_v1.ToDo{
			Title:       fmt.Sprintf("title (%v)", pfx),
			Description: fmt.Sprintf("description (%v)", pfx),
			Reminder:    reminder,
		},
	}

	res1, err := client.Create(ctx, &req1)
	if err != nil {
		log.Fatalf("create todo failed: %v", err)
	}
	log.Printf("create result: %v\n\n", res1)

	id := res1.Id

	// read todo
	req2 := api_v1.ReadRequest{
		Api: apiVersion,
		Id:  id,
	}
	res2, err := client.Read(ctx, &req2)
	if err != nil {
		log.Fatalf("Read failed: %v", err)
	}
	log.Printf("read result: %v\n\n", res2)

	// update
	req3 := api_v1.UpdateRequest{
		Api: apiVersion,
		ToDo: &api_v1.ToDo{
			Id:          res2.ToDo.Id,
			Title:       res2.ToDo.Title,
			Description: res2.ToDo.Description + " updated",
			Reminder:    res2.ToDo.Reminder,
		},
	}
	res3, err := client.Update(ctx, &req3)
	if err != nil {
		log.Fatalf("update failed: %v", err)
	}
	log.Printf("update result: %v\n\n", res3)

	// read all
	req4 := api_v1.ReadAllRequest{
		Api: apiVersion,
	}
	res4, err := client.ReadAll(ctx, &req4)
	if err != nil {
		log.Fatalf("ReadAll failed: %v", err)
	}
	log.Printf("ReadAll result: %v\n\n", res4)

	// Delete todo
	req5 := api_v1.DeleteRequest{
		Api: apiVersion,
		Id:  id,
	}
	res5, err := client.Delete(ctx, &req5)
	if err != nil {
		log.Fatalf("Delete failed: %v", err)
	}
	log.Printf("Delete result: %v\n\n", res5)

	log.Println("done test api_v1")
}

func testingAPI_V2_ServiceA(ctx context.Context, client api_v2.ServiceAClient) {
	log.Println("starting test api_v2 service....")

	msg := &api_v2.MessagePing{
		Timestamp: 111,
	}

	// Anything linked to this variable will transmit request headers.
	md := metadata.New(map[string]string{"x-request-id": "req-service-a"})
	ctx = metadata.NewOutgoingContext(ctx, md)

	// Anything linked to this variable will fetch response headers.
	var header metadata.MD

	reply, err := client.Ping(ctx, msg, grpc.Header(&header))
	if err != nil {
		log.Fatalf("Ping failed: %v", err)
	}
	xrid := header.Get("x-response-id")
	if len(xrid) == 0 {
		log.Fatal("Ping missing 'x-response-id' header")
	}
	if strings.Trim(xrid[0], " ") == "" {
		log.Fatal("Ping empty 'x-response-id' header")
	}
	log.Printf("Ping response: %+v with 'x-response-id': %+v\n", reply, xrid[0])

	reply, err = client.Post(ctx, msg, grpc.Header(&header))
	if err != nil {
		log.Fatalf("Post failed: %v", err)
	}
	xrid = header.Get("x-response-id")
	if len(xrid) == 0 {
		log.Fatal("Post missing 'x-response-id' header")
	}
	if strings.Trim(xrid[0], " ") == "" {
		log.Fatal("Post empty 'x-response-id' header")
	}
	log.Printf("Post response: %+v with 'x-response-id': %+v\n", reply, xrid[0])

	log.Println("done test api_v2 service")
}

func testingAPI_V2_ServiceExtra(ctx context.Context, client api_v2.ServiceExtraClient) {
	log.Println("starting test api_v2 extra service....")

	// Anything linked to this variable will transmit request headers.
	md := metadata.New(map[string]string{"x-request-id": "req-service-extra"})
	ctx = metadata.NewOutgoingContext(ctx, md)

	// incoming response header
	// header, ok := metadata.FromIncomingContext(ctx)

	// grpc stream

	// response stream
	msg := &api_v2.StreamingMessagePing{
		Timestamp:    222,
		MessageCount: 2,
	}
	streamResp, err := client.StreamingPing(ctx, msg)
	if err != nil {
		log.Fatalf("Ping streaming failed: %v", err)
	}
	if header, ok := streamResp.Header(); ok == nil {
		if v, ok := header["error"]; ok {
			log.Fatalf("Ping streaming header error: %v", v)
		}
		log.Printf("Ping streaming headers: %+v\n", header)
	}
	if md := streamResp.Trailer(); md != nil {
		log.Printf("Ping streaming trailer: %+v\n", md)
	}
	for {
		reply, err := streamResp.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Fatalf("Ping streaming received failed: %v", err)
		}
		log.Println("Ping streaming resp:", reply)
	}

	// request stream
	streamReq, err := client.StreamingPost(ctx)
	if err != nil {
		log.Fatalf("Post streaming failed: %v", err)
	}
	var wg sync.WaitGroup
	wg.Add(1)
	go func(wg *sync.WaitGroup) {
		defer wg.Done()
		var count int32
		for {
			if err := streamReq.Send(&api_v2.StreamingMessagePing{
				Timestamp: 333,
			}); err != nil {
				log.Fatalf("Post streaming sended failed: %v", err)
			}
			// time.Sleep(time.Second)
			count++
			if count >= 2 {
				if err := streamReq.CloseSend(); err != nil {
					log.Fatalf("Post streaming close send failed: %v", err)
				}
				return
			}
		}
	}(&wg)
	wg.Wait()
	if header, ok := streamReq.Header(); ok == nil {
		if v, ok := header["error"]; ok {
			log.Fatalf("Post streaming header error: %v", v)
		}
		log.Printf("Post streaming headers: %+v\n", header)
	}
	if reply, err := streamReq.CloseAndRecv(); err != nil {
		log.Fatalf("Post streaming received failed: %v", err)
	} else {
		log.Println("Post streaming received:", reply)
	}
	if md := streamReq.Trailer(); md != nil {
		log.Printf("Post streaming trailer: %+v\n", md)
	}

	// duplex stream
	streamDuplex, err := client.DuplexStreamingPing(ctx)
	if err != nil {
		log.Fatalf("Duplex streaming failed: %v", err)
	}
	go func() {
		count := int32(1)
		for {
			if err := streamDuplex.Send(&api_v2.StreamingMessagePing{
				Timestamp:    444,
				MessageCount: count,
			}); err != nil {
				log.Fatalf("Duplex streaming sended failed: %v", err)
			}
			// time.Sleep(time.Second)
			count++
			if count >= 3 {
				if err := streamDuplex.CloseSend(); err != nil {
					log.Fatalf("Duplex streaming close send failed: %v", err)
				}
				return
			}
		}
	}()
	// receive response header
	if header, ok := streamDuplex.Header(); ok == nil {
		if v, ok := header["error"]; ok {
			log.Fatalf("Duplex streaming header error: %v", v)
		}
		log.Printf("Duplex streaming headers: %+v\n", header)
	}
	for {
		reply, err := streamDuplex.Recv()
		if err == io.EOF {
			if md := streamDuplex.Trailer(); md != nil {
				log.Printf("Duplex streaming trailer: %+v\n", md)
			}
			break
		}
		if err != nil {
			log.Fatalf("Duplex streaming received failed: %v", err)
		}
		log.Println("Duplex streaming resp:", reply)
	}

	log.Println("done test api_v2 extra service")
}

var (
	address = flag.String("grpc-host", "", "gRPC server in format host:port")
)

// remaining single connection to grpc server
func singleConnection() {

	conn, err := grpc.Dial(*address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	// api v1
	// client_v1 := api_v1.NewToDoServiceClient(conn)

	// api v2
	client_v2_serviceA := api_v2.NewServiceAClient(conn)
	client_v2_serviceExtra := api_v2.NewServiceExtraClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// testing
	// testingAPI_V1(ctx, client_v1)
	testingAPI_V2_ServiceA(ctx, client_v2_serviceA)
	testingAPI_V2_ServiceExtra(ctx, client_v2_serviceExtra)

}

// using pool connection with grpcpool
func poolConnections() {
	maxPoolSize := 100
	timeOut := 10

	managerClient := bridge.NewManagerClient(maxPoolSize, timeOut)
	if managerClient == nil {
		log.Fatalf("[Main] Create new manager client failed")
	}

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

	if resp, err := client.Post(timestamp); err != nil {
		log.Fatalf("[Test] Post error: %+v", err)
	} else {
		log.Printf("[Test] Post resp: %+v\n", resp)
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

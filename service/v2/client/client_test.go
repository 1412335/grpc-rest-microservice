package client

import (
	"context"
	"fmt"
	"io"
	"log"
	"strings"
	"sync"
	"testing"
	"time"

	api_v1 "github.com/1412335/grpc-rest-microservice/pkg/api/v1"
	api_v2 "github.com/1412335/grpc-rest-microservice/pkg/api/v2/grpc-gateway/gen"

	"google.golang.org/protobuf/types/known/timestamppb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const (
	apiVersion = "v1"
	address    = "localhost:9090"
)

// remaining single connection to grpc server
func TestSingleConnection(*testing.T) {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	// api v1
	clientV1 := api_v1.NewToDoServiceClient(conn)

	// api v2
	clientV2ServiceA := api_v2.NewServiceAClient(conn)
	clientV2ServiceExtra := api_v2.NewServiceExtraClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// testing
	testingAPIV1(ctx, clientV1)
	testingAPIV2ServiceA(ctx, clientV2ServiceA)
	testingAPIV2ServiceExtra(ctx, clientV2ServiceExtra)
}

func testingAPIV1(ctx context.Context, client api_v1.ToDoServiceClient) {
	log.Println("starting test api_v1....")

	t := time.Now().In(time.UTC)
	reminder := timestamppb.New(t)
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

func testingAPIV2ServiceA(ctx context.Context, client api_v2.ServiceAClient) {
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

func testingAPIV2ServiceExtra(ctx context.Context, client api_v2.ServiceExtraClient) {
	log.Println("starting test api_v2 extra service....")

	// Anything linked to this variable will transmit request headers.
	md := metadata.New(map[string]string{"x-request-id": "req-service-extra"})
	ctx = metadata.NewOutgoingContext(ctx, md)

	// incoming response header
	// header, ok := metadata.FromIncomingContext(ctx)

	// grpc stream
	serverStream(ctx, client)
	clientStream(ctx, client)
	duplexStream(ctx, client)

	log.Println("done test api_v2 extra service")
}

func serverStream(ctx context.Context, client api_v2.ServiceExtraClient) {
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
		reply, e := streamResp.Recv()
		if e != nil {
			if e == io.EOF {
				break
			}
			log.Fatalf("Ping streaming received failed: %v", e)
		}
		log.Println("Ping streaming resp:", reply)
	}
}

func clientStream(ctx context.Context, client api_v2.ServiceExtraClient) {
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
			if e := streamReq.Send(&api_v2.StreamingMessagePing{
				Timestamp: 333,
			}); e != nil {
				log.Fatalf("Post streaming sended failed: %v", e)
			}
			time.Sleep(time.Second)
			count++
			if count >= 2 {
				if e := streamReq.CloseSend(); e != nil {
					log.Fatalf("Post streaming close send failed: %v", e)
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
	if reply, e := streamReq.CloseAndRecv(); e != nil {
		log.Fatalf("Post streaming received failed: %v", e)
	} else {
		log.Println("Post streaming received:", reply)
	}
	if md := streamReq.Trailer(); md != nil {
		log.Printf("Post streaming trailer: %+v\n", md)
	}
}

func duplexStream(ctx context.Context, client api_v2.ServiceExtraClient) {
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
			time.Sleep(time.Second)
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
}

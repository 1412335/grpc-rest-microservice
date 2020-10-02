package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"sync"
	"time"

	api_v1 "grpc-rest-microservice/pkg/api/v1"
	api_v2 "grpc-rest-microservice/pkg/api/v2/gen/grpc-gateway/gen"

	"github.com/golang/protobuf/ptypes"
	"google.golang.org/grpc"
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

	reply, err := client.Ping(ctx, msg)
	if err != nil {
		log.Fatalf("Ping failed: %v", err)
	}
	log.Println("Ping response:", reply)

	reply, err = client.Post(ctx, msg)
	if err != nil {
		log.Fatalf("Post failed: %v", err)
	}
	log.Println("Post response:", reply)

	log.Println("done test api_v2 service")
}

func testingAPI_V2_ServiceExtra(ctx context.Context, client api_v2.ServiceExtraClient) {
	log.Println("starting test api_v2 extra service....")

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
				// streamReq.Send(nil)
				return
			}
		}
	}(&wg)
	wg.Wait()
	if reply, err := streamReq.CloseAndRecv(); err != nil {
		log.Fatalf("Post streaming received failed: %v", err)
	} else {
		log.Println("Post streaming received:", reply)
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
			time.Sleep(time.Second)
			count++
			if count >= 3 {
				// streamDuplex.Send(nil)
				return
			}
		}
	}()
	for {
		reply, err := streamDuplex.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Fatalf("Duplex streaming received failed: %v", err)
		}
		log.Println("Duplex streaming resp:", reply)
	}

	log.Println("done test api_v2 extra service")
}

var (
	address = flag.String("grpc-host", "", "gRPC server in format host:port")
)

func main() {
	flag.Parse()

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

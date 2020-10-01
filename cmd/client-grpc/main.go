package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	api_v1 "github.com/1412335/grpc-rest-microservice/pkg/api/v1"
	"github.com/golang/protobuf/ptypes"
	"google.golang.org/grpc"
)

const (
	apiVersion = "v1"
)

func main() {
	address := flag.String("server", "", "gRPC server in format host:port")
	flag.Parse()

	conn, err := grpc.Dial(*address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	c := api_v1.NewToDoServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

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

	res1, err := c.Create(ctx, &req1)
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
	res2, err := c.Read(ctx, &req2)
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
	res3, err := c.Update(ctx, &req3)
	if err != nil {
		log.Fatalf("update failed: %v", err)
	}
	log.Printf("update result: %v\n\n", res3)

	// read all
	req4 := api_v1.ReadAllRequest{
		Api: apiVersion,
	}
	res4, err := c.ReadAll(ctx, &req4)
	if err != nil {
		log.Fatalf("ReadAll failed: %v", err)
	}
	log.Printf("ReadAll result: %v\n\n", res4)

	// Delete todo
	req5 := api_v1.DeleteRequest{
		Api: apiVersion,
		Id:  id,
	}
	res5, err := c.Delete(ctx, &req5)
	if err != nil {
		log.Fatalf("Delete failed: %v", err)
	}
	log.Printf("Delete result: %v\n\n", res5)

}

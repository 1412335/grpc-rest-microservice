package main

import (
	"fmt"
	"os"

	cmd "grpc-rest-microservice/pkg/cmd/server"
)

func main() {
	if err := cmd.RunServerV2(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

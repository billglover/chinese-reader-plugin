package main

import (
	"fmt"
	"log"

	"golang.org/x/net/context"

	"google.golang.org/grpc"

	tm "github.com/billglover/chinese-reader/token_manager"
)

func main() {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())

	conn, err := grpc.Dial(fmt.Sprintf("localhost:8080"), opts...)
	if err != nil {
		log.Fatalf("failed to dial: %v", err)
	}

	client := tm.NewTokenManagerClient(conn)

	tr := tm.TokenRequest{}

	token, err := client.CreateToken(context.Background(), &tr)
	if err != nil {
		log.Printf("get token: {%+v} %s\n", token, "FAILED")
		log.Printf("unable to get token: %v", grpc.ErrorDesc(err))
		return
	}
	log.Printf("get token: {%+v} %s\n", token, "SUCCESS")
}

package token_service

import (
	"fmt"
	"log"
	"net"

	"golang.org/x/net/context"

	tm "github.com/billglover/chinese-reader/token_manager"
	"google.golang.org/grpc"
)

type TokenManagerServer struct{}

func init() {
	var err error
	var lis net.Listener
	var grpcServer *grpc.Server

	lis, err = net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer = grpc.NewServer()
	tm.RegisterTokenManagerServer(grpcServer, newServer())
	grpcServer.Serve(lis)
}

func (tms *TokenManagerServer) CreateToken(context.Context, *tm.TokenRequest) (*tm.Token, error) {
	fmt.Println("CreateToken")
	t := &tm.Token{}
	return t, nil
}

func (tms *TokenManagerServer) ValidateToken(context.Context, *tm.Token) (*tm.Validity, error) {
	fmt.Println("ValidateToken")
	v := &tm.Validity{}
	return v, nil
}

func (tms *TokenManagerServer) DecrementToken(context.Context, *tm.Token) (*tm.Token, error) {
	fmt.Println("DecrementToken")
	t := &tm.Token{}
	return t, nil
}

func newServer() *TokenManagerServer {
	s := new(TokenManagerServer)
	return s
}

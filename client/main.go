package main

import (
	"log"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/0jk6/tunnel/proto"
)

func main() {

	if len(os.Args) < 3 {
		log.Fatalf("Usage: %s <port> <subdomain>", os.Args[0])
	}

	port := os.Args[1]
	subdomain := os.Args[2]

	conn, err := grpc.NewClient("localhost:12000", grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewTunnelServiceClient(conn)

	handleStream(client, port, subdomain)
}

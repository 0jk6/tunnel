package main

import (
	"log"
	"net"
	"net/http"
	"sync"

	pb "github.com/0jk6/tunnel/proto"
	"google.golang.org/grpc"
)

var addr string = "0.0.0.0:12000"

type Server struct {
	pb.TunnelServiceServer
	streams map[string]pb.TunnelService_ConnectServer
	mu      sync.Mutex
}

func main() {
	lis, err := net.Listen("tcp", addr)

	if err != nil {
		log.Printf("Failed to listen: %v", err)
	}

	s := grpc.NewServer()
	server := Server{streams: make(map[string]pb.TunnelService_ConnectServer)}
	pb.RegisterTunnelServiceServer(s, &server)

	//start the http server
	log.Printf("HTTP server listening at 0.0.0.0:8080")
	http.HandleFunc("/", server.ClientHandler)
	go http.ListenAndServe(":8080", nil)

	//start the grpc server
	log.Printf("gRPC server listening at %s", addr)
	if err := s.Serve(lis); err != nil {
		log.Printf("Failed to serve: %v", err)
	}
}

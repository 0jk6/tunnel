package main

import (
	"log"

	pb "github.com/0jk6/tunnel/proto"
	"google.golang.org/grpc"
)

func (s *Server) Connect(stream grpc.BidiStreamingServer[pb.TunnelMessage, pb.TunnelMessage]) error {
	//receive the first request and store the stream in the map
	req, err := stream.Recv()
	if err != nil {
		return err
	}

	subdomain := req.Subdomain
	if subdomain == "" {
		return nil
	}

	//register the client
	s.mu.Lock()
	s.streams[subdomain] = stream
	s.mu.Unlock()

	log.Println("streams", s.streams)
	log.Println("-----------------")

	//wait until the client disconnects, all the bidirectional stream will happen in the handler.go file
	<-stream.Context().Done()

	//delete the stream from the map when the client disconnects
	s.mu.Lock()
	delete(s.streams, subdomain)
	s.mu.Unlock()

	return nil
}

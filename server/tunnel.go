package main

import (
	"io"
	"log"

	pb "github.com/0jk6/tunnel/proto"
	"google.golang.org/grpc"
)

func (s *Server) Connect(stream grpc.BidiStreamingServer[pb.TunnelMessage, pb.TunnelMessage]) error {
	log.Println("Connected")

	//receive the first request and store the id
	req, err := stream.Recv()
	if err != nil {
		return err
	}

	clientId := req.Id
	if clientId == "" {
		return nil
	}

	s.mu.Lock()
	s.streams[clientId] = stream
	s.mu.Unlock()

	log.Println("streams", s.streams)

	//delete the stream from the map when the client disconnects
	defer func() {
		s.mu.Lock()
		delete(s.streams, clientId)
		s.mu.Unlock()
	}()

	//continue to receive requests and respond to them
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return nil
		}

		if err != nil {
			return err
		}

		err = stream.Send(&pb.TunnelMessage{
			Id:   clientId,
			Body: req.Body,
		})

		if err != nil {
			log.Fatalf("Error while sending data to the client: %v", err)
		}
	}
}

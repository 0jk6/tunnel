package main

import (
	"io"
	"log"
	"net/http"
	"strings"

	pb "github.com/0jk6/tunnel/proto"
)

func (s *Server) WildcardHandler(w http.ResponseWriter, r *http.Request) {
	//whatever requests are coming to this handler should be streamed to the grpc client
	//and the response from the grpc client should be streamed back to the user
	//this is the handler that will be used to handle all requests

	//extract the subdomain from the request
	subdomain := strings.Split(r.Host, ".")[0]

	stream, ok := s.streams[subdomain]

	if !ok {
		w.Write([]byte("stream not found"))
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Fatalf("Error while reading body: %v", err)
		return
	}

	// Copy headers from the incoming request
	headers := make(map[string]string)
	for key, values := range r.Header {
		headers[key] = values[0]
	}

	err = stream.Send(&pb.TunnelMessage{
		Id:      subdomain,
		Body:    body,
		Path:    r.URL.Path,
		Headers: headers,
		Method:  r.Method,
	})

	if err != nil {
		log.Fatalf("Error while sending data to the client: %v", err)
	}

	res, err := stream.Recv()

	if err == io.EOF {
		return
	}

	// Set response headers based on the gRPC response
	for key, value := range res.Headers {
		w.Header().Set(key, value)
	}

	w.WriteHeader(200)

	w.Write(res.Body)
}

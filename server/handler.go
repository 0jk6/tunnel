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
		http.Error(w, "stream not found", http.StatusNotFound)
		return
	}

	var body []byte
	if r.Body != nil {
		defer r.Body.Close()
		var err error
		body, err = io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "failed to read request body", http.StatusInternalServerError)
			log.Fatalf("Error while reading body: %v", err)
			return
		}
	}

	// Copy headers from the incoming request
	headers := make(map[string]string)
	for key, values := range r.Header {
		if len(values) > 0 {
			headers[key] = values[0]
		}
	}

	err := stream.Send(&pb.TunnelMessage{
		Id:      subdomain,
		Body:    body,
		Path:    r.URL.Path,
		Headers: headers,
		Method:  r.Method,
	})

	if err != nil {
		http.Error(w, "failed to send data to gRPC client", http.StatusInternalServerError)
		log.Fatalf("Error while sending data to the gRPC client: %v", err)
		return
	}

	res, err := stream.Recv()
	if err != nil {
		if err == io.EOF {
			return
		}
		http.Error(w, "failed to receive data from gRPC client", http.StatusInternalServerError)
		log.Printf("Error while receiving data from the client: %v", err)
		return
	}

	if res != nil {
		// Set response headers based on the gRPC response
		for key, value := range res.Headers {
			w.Header().Set(key, value)
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(res.Body)
	} else {
		http.Error(w, "empty response from gRPC client", http.StatusInternalServerError)
	}
}

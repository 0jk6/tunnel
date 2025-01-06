package main

import (
	"io"
	"log"
	"net/http"
	"strings"

	pb "github.com/0jk6/tunnel/proto"
)

func (s *Server) WildcardHandler(w http.ResponseWriter, r *http.Request) {
	subdomain := strings.Split(r.Host, ".")[0]

	s.mu.Lock()
	stream, ok := s.streams[subdomain]
	s.mu.Unlock()

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
		Subdomain: subdomain,
		Body:      body,
		Path:      r.URL.Path,
		Headers:   headers,
		Method:    r.Method,
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
		delete(s.streams, subdomain)
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

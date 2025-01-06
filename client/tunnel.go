package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"

	pb "github.com/0jk6/tunnel/proto"
)

func handleStream(client pb.TunnelServiceClient, port, subdomain string) {
	stream, err := client.Connect(context.Background())
	if err != nil {
		log.Fatalf("Failed to create stream: %v", err)
	}

	//send the first request to register the client
	err = stream.Send(&pb.TunnelMessage{
		Subdomain: subdomain,
	})
	if err != nil {
		log.Fatalf("Failed to send: %v", err)
	}

	//channel to wait for the goroutine
	waitc := make(chan struct{})

	//receive the response from the server, process it and send back the response
	go func() {
		for {
			res, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Fatalf("Failed to receive: %v", err)
				break
			}

			// Create a new HTTP request
			url := fmt.Sprintf("http://localhost:%s%s", port, res.Path)
			log.Println(url)
			req, err := http.NewRequest(res.Method, url, bytes.NewReader(res.Body))
			if err != nil {
				log.Printf("Failed to create request: %v", err)
				// break
			}

			// Set headers from the gRPC response
			for key, value := range res.Headers {
				req.Header.Set(key, value)
			}

			// Make the HTTP request
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				log.Printf("Failed to make request: %v", err)
				continue
			}

			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()

			log.Println(string(body))

			// Based on the server's response, send back a response
			err = stream.Send(&pb.TunnelMessage{
				Subdomain: subdomain,
				Body:      body,
			})

			if err != nil {
				log.Fatalf("Failed to send response to the gRPC server: %v", err)
				// break
			}

		}
		close(waitc)
	}()

	<-waitc
}
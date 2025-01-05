package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
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

	conn, err := grpc.NewClient("localhost:12000", grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(1024*1024*50)))

	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewTunnelServiceClient(conn)

	stream, err := client.Connect(context.Background())

	if err != nil {
		log.Fatalf("Failed to create stream: %v", err)
	}

	log.Println(os.Args[1])

	//send the first request to register the client
	err = stream.Send(&pb.TunnelMessage{
		Id: subdomain,
	})

	if err != nil {
		log.Fatalf("Failed to send: %v", err)
	}

	waitc := make(chan struct{})

	//receive the response from the server and send back a response
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

			// log.Println(res.Path)

			// Create a new HTTP request
			url := fmt.Sprintf("http://localhost:%s%s", port, res.Path)
			log.Println(url)
			req, err := http.NewRequest(res.Method, url, nil)
			if err != nil {
				log.Fatalf("Failed to create request: %v", err)
				// break
			}

			// Set headers from the gRPC response
			for key, value := range res.Headers {
				req.Header.Set(key, value)
			}

			// Make the HTTP request
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				log.Fatalf("Failed to make request: %v", err)
				// break
			}

			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()

			// Based on the server's response, send back a response
			err = stream.Send(&pb.TunnelMessage{
				Id:   subdomain,
				Body: body,
			})

			if err != nil {
				log.Fatalf("Failed to send: %v", err)
				// break
			}

		}
		close(waitc)
	}()

	<-waitc
}

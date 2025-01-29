package main

import (
	"bufio"
	"bytes"
	"log"
	"net"
	"net/http"
	"strconv"

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

//tcp server implementation

func startTCPServer(server *Server) {
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatalf("Failed to start TCP server: %v", err)
	}

	defer listener.Close()

	log.Printf("TCP server listening at 0.0.0.0:8080")

	// Accept incoming connections
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Failed to accept connection: %v", err)
			continue
		}
		go handleTCPConnection(conn, server)
	}

	//handle connection
}

func handleTCPConnection(conn net.Conn, server *Server) {
	defer conn.Close()

	//determine if the request is an HTTP request or a raw TCP request
	reader := bufio.NewReader(conn)
	peek, err := reader.Peek(8) //peek the first 8 bytes to determine if it is an HTTP request
	if err != nil {
		log.Printf("Failed to peek: %v", err)
		return
	}

	if bytes.HasPrefix(peek, []byte("GET ")) || bytes.HasPrefix(peek, []byte("POST")) || bytes.HasPrefix(peek, []byte("PUT")) || bytes.HasPrefix(peek, []byte("DELETE")) || bytes.HasPrefix(peek, []byte("PATCH")) || bytes.HasPrefix(peek, []byte("OPTIONS")) {
		//HTTP request
		req, err := http.ReadRequest(reader)
		if err != nil {
			log.Printf("Failed to read request: %v", err)
			return
		}

		tcpResponseWriter := &TCPResponseWriter{conn: conn, header: http.Header{}}
		server.ClientHandler(tcpResponseWriter, req)
	} else {
		//raw TCP request, send this tcp data directly to client over grpc
		//todo: edit client to make tcp requests
		log.Println("Raw TCP request")
	}

}

// following struct should implement the http.ResponseWriter interface
// and we can pass this to the http handler that we have in the handler.go file
type TCPResponseWriter struct {
	conn   net.Conn
	header http.Header
	status int
}

// WriteHeader sends an HTTP response header with the provided status code
func (t *TCPResponseWriter) WriteHeader(statusCode int) {
	t.status = statusCode
	statusLine := "HTTP/1.1 " + strconv.Itoa(statusCode) + " " + http.StatusText(statusCode) + "\r\n"
	t.conn.Write([]byte(statusLine))
	for key, values := range t.header {
		for _, value := range values {
			t.conn.Write([]byte(key + ": " + value + "\r\n"))
		}
	}
	t.conn.Write([]byte("\r\n"))
}

func (t *TCPResponseWriter) Write(b []byte) (int, error) {
	return t.conn.Write(b)
}

// Header returns the header map that will be sent by WriteHeader
func (t *TCPResponseWriter) Header() http.Header {
	return t.header
}

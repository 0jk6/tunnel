# Tunnel

An attempt to build a clone of Ngrok. While it may be a bit buggy, it works and demonstrates the basic functionality of tunneling HTTP requests.  

---

## Features
- **gRPC Bi-Directional Streams**: Enables seamless communication between the client and server.
- **Subdomain Mapping**: Access your HTTP server on a custom subdomain like `subdomain.localhost:8080`.
- **Work in Progress**: The code is a mess.

---

### Prerequisites
Ensure you have the following installed:
- [Go](https://golang.org/)
- [Make](https://www.gnu.org/software/make/)
- [Protocol Buffers Compiler (`protoc`)](https://grpc.io/docs/protoc-installation/)

### Building
1. Generate gRPC structures:
   ```bash
   make proto
   ```
2. Install dependencies
    ```bash
    go mod tidy
    ```
3. Build the binaries:
   ```bash
   make build
   ```

### Running
1. Start the **server**:
   ```bash
   ./bin/server
   ```
2. Start the **client** with the desired port and subdomain:
   ```bash
   ./bin/client <port> <subdomain>
   ```

   Example:
   ```bash
   ./bin/client 3000 mysubdomain
   ```

3. Access your HTTP server on:
   ```
   mysubdomain.localhost:8080
   ```

---

## How It Works
- The **client** and **server** communicate using gRPC Bi-Directional streams, facilitating real-time tunneling of requests.

---
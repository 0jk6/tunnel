.PHONY: build proto

proto:
	protoc -Iproto --go_out=. --go_opt=module=github.com/0jk6/tunnel --go-grpc_out=. --go-grpc_opt=module=github.com/0jk6/tunnel proto/tunnel.proto

build:
	go build -o bin/server server/*.go
	go build -o bin/client client/*.go

run-server:
	go build -o bin/server server/*.go
	./bin/server

run-client:
	go build -o bin/client client/*.go
	./bin/client
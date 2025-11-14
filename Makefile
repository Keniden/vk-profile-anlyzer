all: build

run:
	go run ./cmd/server

grpc:
	go run ./cmd/grpcserver

test:
	go test ./...

lint:
	golangci-lint run ./...

build:
	go build ./cmd/server

docker-build:
	docker build -t inteam-api .

docker-up:
	docker-compose up --build

.PHONY: all run grpc test lint build docker-build docker-up

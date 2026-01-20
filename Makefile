.PHONY: build run test test-race test-coverage clean lint fmt deps

build:
	go build -o bin/server cmd/server/main.go

run:
	go run cmd/server/main.go

test:
	go test ./... -v

test-race:
	go test ./... -v -race

test-coverage:
	go test ./... -v -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html

clean:
	rm -rf bin/
	rm -f coverage.out coverage.html

lint:
	golangci-lint run

fmt:
	go fmt ./...

deps:
	go mod download
	go mod tidy
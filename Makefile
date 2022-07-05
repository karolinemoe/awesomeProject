build:
	go build -ldflags "-X main.gitSHA=$(shell git rev-parse HEAD)" main.go

run:
	./main.go

deps:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

lint:
	$(GOPATH)/bin/golangci-lint run --config .golangci.yml
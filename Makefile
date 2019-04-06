GO_SRC := $(shell find . -name "*.go")

.PHONY: all clean fmt test

all: localnews

localnews: $(GO_SRC)
	go build -o bin/localnews cmd/main.go

fmt:
	go fmt ./...

test: $(PROTOBUF_OUT)
	go test ./...

coverage: $(PROTOBUF_OUT)
	go test ./... -coverprofile=coverage.out && go tool cover -html=coverage.out

clean:
	rm -rf bin/*

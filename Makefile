PKG := github.com/opendexnetwork/opendex-docker-api

GO_BIN := ${GOPATH}/bin

GOBUILD := go build -v

VERSION := local
COMMIT := $(shell git rev-parse HEAD)
ifeq ($(OS),Windows_NT)
	TIMESTAMP := $(shell powershell.exe scripts\get_timestamp.ps1)
else
	TIMESTAMP := $(shell date +%s)
endif
LDFLAGS := -ldflags "-w -s \
-X $(PKG)/build.Version=$(VERSION) \
-X $(PKG)/build.GitCommit=$(COMMIT) \
-X $(PKG)/build.Timestamp=$(TIMESTAMP)"

default: build

.PHONY: build
build:
	$(GOBUILD) $(LDFLAGS) ./cmd/proxy

.PHONY: proto
proto:
	echo "Compiling opendexd gRPC protocols"
	protoc -I service/opendexd/proto opendexrpc.proto --go_out=plugins=grpc:service/opendexd/opendexrpc
	echo "Compiling lnd gRPC protocols"
	protoc -I service/lnd/proto rpc.proto --go_out=plugins=grpc:service/lnd/lnrpc
	echo "Compiling boltz gRPC protocols"
	protoc -I service/boltz/proto boltzrpc.proto --go_out=plugins=grpc:service/boltz/boltzrpc

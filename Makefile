.PHONY: phistage phistagecli test grpc

PKG = github.com/projecteru2/phistage
VERSIONPKG = $(PKG)/cmd/phistage/version

REVISION := $(shell git rev-parse HEAD)
BUILTAT := $(shell date +%Y-%m-%dT%H:%M:%S)

TAGCOMMIT := $(git rev-list --tags --max-count=1)
ifeq ($(TAGCOMMIT),)
	VERSION := cesium
else
	VERSION := $(shell git describe --tags $(TAGCOMMIT))
endif

GO_LDFLAGS ?= -s -X $(VERSIONPKG).REVISION=$(REVISION) \
                 -X $(VERSIONPKG).BUILTAT=$(BUILTAT) \
				 -X $(VERSIONPKG).VERSION=$(VERSION)

all: phistage phistagecli

phistage:
	mkdir -p bin
	go build -ldflags "$(GO_LDFLAGS)" -o bin/phistage ./cmd/phistage

phistagecli:
	mkdir -p bin
	go build -ldflags "$(GO_LDFLAGS)" -o bin/phistagecli ./cmd/phistagecli

test:
	go test -v -cover -count=1 ./...

grpc:
	protoc --go_out=plugins=grpc:./apiserver/grpc/proto \
		   --go_opt=paths=source_relative \
		   --proto_path=./apiserver/grpc/proto phistage.proto

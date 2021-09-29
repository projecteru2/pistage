.PHONY: pistage pistagecli test grpc

PKG = github.com/projecteru2/pistage
VERSIONPKG = $(PKG)/cmd/pistage/version

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

all: pistage pistagecli

pistage:
	mkdir -p bin
	go build -ldflags "$(GO_LDFLAGS)" -o bin/pistage ./cmd/pistage

pistagecli:
	mkdir -p bin
	go build -ldflags "$(GO_LDFLAGS)" -o bin/pistagecli ./cmd/pistagecli

test:
	go test -v -cover -count=1 ./...

grpc:
	protoc --go_out=. --go-grpc_out=. \
		   --go_opt=paths=source_relative --go-grpc_opt=paths=source_relative \
		   ./apiserver/grpc/proto/pistage.proto

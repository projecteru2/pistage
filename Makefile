.PHONY: phistage test

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

all: phistage

phistage:
	mkdir -p bin
	go build -ldflags "$(GO_LDFLAGS)" -o bin/phistage ./cmd/phistage
test:
	go test -v -cover -count=1 ./...

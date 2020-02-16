NS := github.com/projecteru2/aa
BUILD := go build -race
TEST := go test -count=1 -race -cover

LDFLAGS += -X "$(NS)/ver.Git=$(shell git rev-parse HEAD)"
LDFLAGS += -X "$(NS)/ver.Compile=$(shell go version)"
LDFLAGS += -X "$(NS)/ver.Date=$(shell date +'%F %T %z')"

PKGS := $$(go list ./...)

.PHONY: all test build

default: build

build: build-server build-cli

build-server:
	echo

build-cli:
	$(BUILD) -ldflags '$(LDFLAGS)' -o bin/aa-cli cli/cli.go

lint: fmt
	golint $(PKGS)
	golangci-lint run

fmt: vet
	gofmt -s -w $$(find . -iname '*.go' | grep -v -P '\./vendor/')

vet:
	go vet $(PKGS)

deps:
	GO111MODULE=on go mod download
	GO111MODULE=on go mod vendor

test:
ifdef RUN
	$(TEST) -v -run='${RUN}' $(PKGS)
else
	$(TEST) $(PKGS)
endif

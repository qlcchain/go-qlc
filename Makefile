.PHONY: clean lint snapshot release
.PHONY: build build-test
.PHONY: deps

# Check for required command tools to build or stop immediately
EXECUTABLES = git go find pwd
K := $(foreach exec,$(EXECUTABLES),\
        $(if $(shell which $(exec)),some string,$(error "No $(exec) in PATH)))

# server
VERSION ?= 1.2.6.6
BINARY = gqlc
MAIN = cmd/main.go

BUILDDIR = build
GITREV = $(shell git rev-parse --short HEAD)
BUILDTIME = $(shell date +'%FT%TZ%z')
GO_BUILDER_VERSION=v1.13.1

deps:
	go get -u github.com/golangci/golangci-lint/cmd/golangci-lint

build:
	go build -ldflags "-X github.com/qlcchain/go-qlc/chain.Version=${VERSION} \
		-X github.com/qlcchain/go-qlc/chain.GitRev=${GITREV} \
        -X github.com/qlcchain/go-qlc/chain.BuildTime=${BUILDTIME} \
        -X github.com/qlcchain/go-qlc/chain.Mode=MainNet" -v -i -o $(shell pwd)/$(BUILDDIR)/$(BINARY) $(shell pwd)/$(MAIN)
	@echo "Build $(BINARY) done."
	@echo "Run \"$(shell pwd)/$(BUILDDIR)/$(BINARY)\" to start $(BINARY)."

build-test:
	go build -tags "testnet" -ldflags "-X github.com/qlcchain/go-qlc/chain.Version=${VERSION} \
		-X github.com/qlcchain/go-qlc/chain.GitRev=${GITREV} \
		-X github.com/qlcchain/go-qlc/chain.BuildTime=${BUILDTIME} \
		-X github.com/qlcchain/go-qlc/chain.Mode=TestNet" -v -i -o $(shell pwd)/$(BUILDDIR)/$(BINARY) $(shell pwd)/$(MAIN)
	@echo "Build testnet $(BINARY) done."
	@echo "Run \"$(BUILDDIR)/$(BINARY)\" to start $(BINARY)."

clean:
	rm -rf $(shell pwd)/$(BUILDDIR)/

snapshot:
	docker run --rm --privileged \
    -v $(CURDIR):/go-qlc \
    -v /var/run/docker.sock:/var/run/docker.sock \
	-v $(GOPATH)/src:/go/src \
    -w /go-qlc \
    goreng/golang-cross:$(GO_BUILDER_VERSION) \
    goreleaser --snapshot --rm-dist

release:
	docker run --rm --privileged \
	-e GITHUB_TOKEN=$(GITHUB_TOKEN) \
    -v $(CURDIR):/go-qlc \
    -v /var/run/docker.sock:/var/run/docker.sock \
	-v $(GOPATH)/src:/go/src \
    -w /go-qlc \
    goreng/golang-cross:$(GO_BUILDER_VERSION) \
    goreleaser --rm-dist

lint: 
	golangci-lint run --fix

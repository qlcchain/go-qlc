.PHONY: clean lint changelog snapshot release
.PHONY: build build-test
.PHONY: deps

# Check for required command tools to build or stop immediately
EXECUTABLES = git go find pwd
K := $(foreach exec,$(EXECUTABLES),\
        $(if $(shell which $(exec)),some string,$(error "No $(exec) in PATH)))

# server
VERSION ?= $(shell git describe --abbrev=0 --tags)
BINARY = gqlc
MAIN = cmd/main.go

BUILDDIR = build
GITREV = $(shell git rev-parse --short HEAD)
BUILDTIME = $(shell date +'%FT%TZ%z')
GO_BUILDER_VERSION=v1.13.1

deps:
	go get -u github.com/golangci/golangci-lint/cmd/golangci-lint
	go get -u github.com/git-chglog/git-chglog/cmd/git-chglog

build:
	go build -ldflags "-X github.com/qlcchain/go-qlc/chain/version.Version=${VERSION} \
		-X github.com/qlcchain/go-qlc/chain/version.GitRev=${GITREV} \
        -X github.com/qlcchain/go-qlc/chain/version.BuildTime=${BUILDTIME} \
        -X github.com/qlcchain/go-qlc/chain/version.Mode=MainNet" -i -o $(shell pwd)/$(BUILDDIR)/$(BINARY) $(shell pwd)/$(MAIN)
	@echo "Build $(BINARY) done."
	@echo "Run \"$(shell pwd)/$(BUILDDIR)/$(BINARY)\" to start $(BINARY)."

build-test:
	go build -tags "testnet" -ldflags "-X github.com/qlcchain/go-qlc/chain/version.Version=${VERSION} \
		-X github.com/qlcchain/go-qlc/chain/version.GitRev=${GITREV} \
		-X github.com/qlcchain/go-qlc/chain/version.BuildTime=${BUILDTIME} \
		-X github.com/qlcchain/go-qlc/chain/version.Mode=TestNet" -i -o $(shell pwd)/$(BUILDDIR)/$(BINARY) $(shell pwd)/$(MAIN)
	@echo "Build testnet $(BINARY) done."
	@echo "Run \"$(BUILDDIR)/$(BINARY)\" to start $(BINARY)."

clean:
	rm -rf $(shell pwd)/$(BUILDDIR)/

changelog:
	git-chglog $(VERSION) > CHANGELOG.md

snapshot:
	docker run --rm --privileged \
		-v $(CURDIR):/go-qlc \
		-v /var/run/docker.sock:/var/run/docker.sock \
		-v $(GOPATH)/src:/go/src \
		-w /go-qlc \
		goreng/golang-cross:$(GO_BUILDER_VERSION) \
		goreleaser --snapshot --rm-dist

release: changelog
	docker run --rm --privileged \
		-e GITHUB_TOKEN=$(GITHUB_TOKEN) \
		-v $(CURDIR):/go-qlc \
		-v /var/run/docker.sock:/var/run/docker.sock \
		-v $(GOPATH)/src:/go/src \
		-w /go-qlc \
		goreng/golang-cross:$(GO_BUILDER_VERSION) \
		goreleaser --rm-dist --release-notes=CHANGELOG.md

lint: 
	golangci-lint run --fix

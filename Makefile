.PHONY: clean lint changelog snapshot release
.PHONY: build build-test
.PHONY: deps
.PHONY: mockery

# Check for required command tools to build or stop immediately
EXECUTABLES = git go find pwd
K := $(foreach exec,$(EXECUTABLES),\
        $(if $(shell which $(exec)),some string,$(error "No $(exec) in PATH")))

# server
VERSION ?= $(shell git describe --tags `git rev-list --tags --max-count=1`)
BINARY = gqlc
MAIN = cmd/main.go

BUILDDIR = build
GITREV = $(shell git rev-parse --short HEAD)
BUILDTIME = $(shell date +'%FT%TZ%z')
GO_BUILDER_VERSION=v1.16.3

deps:
	go get -u github.com/golangci/golangci-lint/cmd/golangci-lint
	go get -u github.com/git-chglog/git-chglog/cmd/git-chglog
	go get -u golang.org/x/tools/cmd/goimports
	go get -u github.com/vektra/mockery/.../

build:
	go build -ldflags "-X github.com/qlcchain/go-qlc/chain/version.Version=${VERSION} \
		-X github.com/qlcchain/go-qlc/chain/version.GitRev=${GITREV} \
        -X github.com/qlcchain/go-qlc/chain/version.BuildTime=${BUILDTIME} \
        -X github.com/qlcchain/go-qlc/chain/version.Mode=MainNet" -o $(shell pwd)/$(BUILDDIR)/$(BINARY) $(shell pwd)/$(MAIN)
	@echo "Build $(BINARY) done."
	@echo "Run \"$(shell pwd)/$(BUILDDIR)/$(BINARY)\" to start $(BINARY)."

build-test:
	go build -tags "testnet" -ldflags "-X github.com/qlcchain/go-qlc/chain/version.Version=${VERSION} \
		-X github.com/qlcchain/go-qlc/chain/version.GitRev=${GITREV} \
		-X github.com/qlcchain/go-qlc/chain/version.BuildTime=${BUILDTIME} \
		-X github.com/qlcchain/go-qlc/chain/version.Mode=TestNet" -o $(shell pwd)/$(BUILDDIR)/$(BINARY) $(shell pwd)/$(MAIN)
	@echo "Build testnet $(BINARY) done."
	@echo "Run \"$(BUILDDIR)/$(BINARY)\" to start $(BINARY)."

clean:
	rm -rf $(shell pwd)/$(BUILDDIR)/

changelog:
	git-chglog $(VERSION) > CHANGELOG.md
	@cat assets/footer.txt >> CHANGELOG.md

snapshot:
	docker run --rm --privileged \
		-e PRIVATE_KEY=$(PRIVATE_KEY) \
		-v $(CURDIR):/go-qlc \
		-v /var/run/docker.sock:/var/run/docker.sock \
		-v $(GOPATH)/src:/go/src \
		-w /go-qlc \
		goreng/golang-cross:$(GO_BUILDER_VERSION) --snapshot --rm-dist
		
release: changelog
	docker run --rm --privileged \
		-e GITHUB_TOKEN=$(GITHUB_TOKEN) \
		-e PRIVATE_KEY=$(PRIVATE_KEY) \
		-v $(CURDIR):/go-qlc \
		-v /var/run/docker.sock:/var/run/docker.sock \
		-v $(GOPATH)/src:/go/src \
		-w /go-qlc \
		goreng/golang-cross:$(GO_BUILDER_VERSION) --rm-dist --release-notes=CHANGELOG.md

lint: 
	golangci-lint run --fix

mockery:
	 mockery -name=Store -dir=./ledger -output ./mock/mocks/

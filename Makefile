.PHONY: all clean
.PHONY: gqlc-server gqlc-server-test
.PHONY: gqlc-client
.PHONY: deps

# Check for required command tools to build or stop immediately
EXECUTABLES = git go find pwd docker
K := $(foreach exec,$(EXECUTABLES),\
        $(if $(shell which $(exec)),some string,$(error "No $(exec) in PATH)))

GO ?= latest

# server
SERVERVERSION = 1.0.4
SERVERBINARY = gqlc
SERVERTESTBINARY = gqlct
SERVERMAIN = cmd/server/main.go

# client
CLIENTVERSION = 1.0.4
CLIENTBINARY = gqlcc
CLIENTMAIN = cmd/client/main.go

BUILDDIR = $(shell pwd)/build
GITREV = $(shell git rev-parse --short HEAD)
BUILDTIME = $(shell date +'%Y-%m-%d_%T')

MAINLDFLAGS="-X github.com/qlcchain/go-qlc/cmd/server/commands.Version=${SERVERVERSION} \
	-X github.com/qlcchain/go-qlc/cmd/server/commands.GitRev=${GITREV} \
	-X github.com/qlcchain/go-qlc/cmd/server/commands.BuildTime=${BUILDTIME} \
	-X github.com/qlcchain/go-qlc/cmd/server/commands.Mode=MainNet"

TESTLDFLAGS="-X github.com/qlcchain/go-qlc/cmd/server/commands.Version=${SERVERVERSION} \
	-X github.com/qlcchain/go-qlc/cmd/server/commands.GitRev=${GITREV} \
	-X github.com/qlcchain/go-qlc/cmd/server/commands.BuildTime=${BUILDTIME} \
	-X github.com/qlcchain/go-qlc/cmd/server/commands.Mode=TestNet"

CLIENTLDFLAGS="-X github.com/qlcchain/go-qlc/cmd/client/commands.Version=${CLIENTVERSION} \
	-X github.com/qlcchain/go-qlc/cmd/client/commands.GitRev=${GITREV} \
	-X github.com/qlcchain/go-qlc/cmd/client/commands.BuildTime=${BUILDTIME}" \

deps:
	go get -u golang.org/x/lint/golint
	go get -u github.com/gythialy/xgo
	go get -u github.com/git-chglog/git-chglog/cmd/git-chglog

build:
	go build -tags "mainnet sqlite_userauth" -ldflags $(MAINLDFLAGS) -v -i -o $(BUILDDIR)/$(SERVERBINARY) $(shell pwd)/$(SERVERMAIN)
	@echo "Build server done."
	@echo "Run \"$(BUILDDIR)/$(SERVERBINARY)\" to start $(SERVERTESTBINARY)."
	go build -tags mainnet -ldflags $(CLIENTLDFLAGS) -v -i -o $(BUILDDIR)/$(CLIENTBINARY) $(shell pwd)/$(CLIENTMAIN)
	@echo "Build client done."
	@echo "Run \"$(BUILDDIR)/$(CLIENTBINARY)\" to start $(SERVERTESTBINARY)."

build-test:
	go build -tags "testnet sqlite_userauth" -ldflags $(TESTLDFLAGS) -v -i -o $(BUILDDIR)/$(SERVERBINARY) $(shell pwd)/$(SERVERMAIN)
	@echo "Build test server done."
	@echo "Run \"$(BUILDDIR)/$(SERVERBINARY)\" to start $(SERVERBINARY)."
	go build -tags mainnet -ldflags $(CLIENTLDFLAGS) -v -i -o $(BUILDDIR)/$(CLIENTBINARY) $(shell pwd)/$(CLIENTMAIN)
	@echo "Build test client done."
	@echo "Run \"$(BUILDDIR)/$(CLIENTBINARY)\" to start $(CLIENTBINARY)."

all: gqlc-server gqlc-server-test gqlc-client

clean:
	rm -rf $(BUILDDIR)/

gqlc-server:
	xgo --dest=$(BUILDDIR) --tags="mainnet sqlite_userauth" --ldflags=$(MAINLDFLAGS) --out=$(SERVERBINARY)-v$(SERVERVERSION)-$(GITREV) \
	--targets="windows/*,darwin/amd64,linux/amd64,linux/386,linux/arm64,linux/mips64, linux/mips64le" \
	--pkg=$(SERVERMAIN) .

gqlc-server-test:
	xgo --dest=$(BUILDDIR) --tags="testnet sqlite_userauth" --ldflags=$(TESTLDFLAGS) --out=$(SERVERTESTBINARY)-v$(SERVERVERSION)-$(GITREV) \
	--targets="windows/amd64,darwin/amd64,linux/amd64,linux/386,linux/arm64,linux/mips64, linux/mips64le" \
	--pkg=$(SERVERMAIN) .

gqlc-client:
	xgo --tags=mainnet --dest=$(BUILDDIR) --ldflags=$(CLIENTLDFLAGS) --out=$(CLIENTBINARY)-v$(CLIENTVERSION)-$(GITREV) \
	--targets="windows/amd64,darwin/amd64,linux/amd64" \
	--pkg=$(CLIENTMAIN) .
.PHONY: all clean build build-test confidant confidant-test
.PHONY: gqlc-server gqlc-server-test
.PHONY: deps

# Check for required command tools to build or stop immediately
EXECUTABLES = git go find pwd
K := $(foreach exec,$(EXECUTABLES),\
        $(if $(shell which $(exec)),some string,$(error "No $(exec) in PATH)))

# server
SERVERVERSION ?= 1.2.6.6
SERVERBINARY = gqlc
SERVERTESTBINARY = gqlct
SERVERMAIN = cmd/main.go

BUILDDIR = build
GITREV = $(shell git rev-parse --short HEAD)
BUILDTIME = $(shell date +'%Y-%m-%d_%T')
TARGET=windows-6.0/*,darwin-10.10/amd64,linux/amd64
TARGET_CONFIDANT=linux/arm-7

MAINLDFLAGS="-X github.com/qlcchain/go-qlc/cmd/server/commands.Version=${SERVERVERSION} \
	-X github.com/qlcchain/go-qlc/cmd/server/commands.GitRev=${GITREV} \
	-X github.com/qlcchain/go-qlc/cmd/server/commands.BuildTime=${BUILDTIME} \
	-X github.com/qlcchain/go-qlc/cmd/server/commands.Mode=MainNet"

TESTLDFLAGS="-X github.com/qlcchain/go-qlc/cmd/server/commands.Version=${SERVERVERSION} \
	-X github.com/qlcchain/go-qlc/cmd/server/commands.GitRev=${GITREV} \
	-X github.com/qlcchain/go-qlc/cmd/server/commands.BuildTime=${BUILDTIME} \
	-X github.com/qlcchain/go-qlc/cmd/server/commands.Mode=TestNet"

deps:
	go get -u golang.org/x/lint/golint
	go get -u github.com/gythialy/xgo
	go get -u github.com/git-chglog/git-chglog/cmd/git-chglog

confidant:
	CGO_ENABLED=1 CC=/opt/gcc-linaro-5.3.1-2016.05-x86_64_arm-linux-gnueabihf/bin/arm-linux-gnueabihf-gcc GOARCH=arm GOARM=7 \
	GO111MODULE=on go build -tags "confidant" -ldflags $(MAINLDFLAGS) -v -i -o $(shell pwd)/$(BUILDDIR)/$(SERVERBINARY) $(shell pwd)/$(SERVERMAIN)
	@echo "Build $(SERVERBINARY) done."
	@echo "Run \"$(shell pwd)/$(BUILDDIR)/$(SERVERBINARY)\" to start $(SERVERBINARY)."

confidant-test:
	CGO_ENABLED=1 CC=/opt/gcc-linaro-5.3.1-2016.05-x86_64_arm-linux-gnueabihf/bin/arm-linux-gnueabihf-gcc GOARCH=arm GOARM=7 \
	GO111MODULE=on go build -tags "confidant testnet" -ldflags $(MAINLDFLAGS) -v -i -o $(shell pwd)/$(BUILDDIR)/$(SERVERBINARY) $(shell pwd)/$(SERVERMAIN)
	@echo "Build $(SERVERBINARY) done."
	@echo "Run \"$(shell pwd)/$(BUILDDIR)/$(SERVERBINARY)\" to start $(SERVERBINARY)."

build:
	GO111MODULE=on go build -ldflags $(MAINLDFLAGS) -v -i -o $(shell pwd)/$(BUILDDIR)/$(SERVERBINARY) $(shell pwd)/$(SERVERMAIN)
	@echo "Build $(SERVERBINARY) done."
	@echo "Run \"$(shell pwd)/$(BUILDDIR)/$(SERVERBINARY)\" to start $(SERVERBINARY)."

build-test:
	GO111MODULE=on go build -tags "testnet" -ldflags $(TESTLDFLAGS) -v -i -o $(shell pwd)/$(BUILDDIR)/$(SERVERBINARY) $(shell pwd)/$(SERVERMAIN)
	@echo "Build test server done."
	@echo "Run \"$(BUILDDIR)/$(SERVERBINARY)\" to start $(SERVERBINARY)."

mining:
	GO111MODULE=on go build -tags "testnet" -ldflags $(TESTLDFLAGS) -v -i -o $(shell pwd)/$(BUILDDIR)/gqlc-miner $(shell pwd)/cmd/miner/
	GO111MODULE=on go build -tags "testnet" -ldflags $(TESTLDFLAGS) -v -i -o $(shell pwd)/$(BUILDDIR)/gqlc-stratum $(shell pwd)/cmd/stratum/
	@echo "Build mining done."

all: gqlc-server gqlc-server-test gqlc-client

clean:
	rm -rf $(shell pwd)/$(BUILDDIR)/

gqlc-server:
	xgo --dest=$(BUILDDIR) --ldflags=$(MAINLDFLAGS) --out=$(SERVERBINARY)-v$(SERVERVERSION)-$(GITREV) \
    --targets=$(TARGET) --pkg=$(SERVERMAIN) .
	xgo --dest=$(BUILDDIR) --tags="confidant" --ldflags=$(MAINLDFLAGS) --out=$(SERVERBINARY)-confidant-v$(SERVERVERSION)-$(GITREV) \
	--targets=$(TARGET_CONFIDANT) --pkg=$(SERVERMAIN) .

gqlc-server-test:
	xgo --dest=$(BUILDDIR) --tags="testnet" --ldflags=$(TESTLDFLAGS) --out=$(SERVERTESTBINARY)-v$(SERVERVERSION)-$(GITREV) \
	--targets=$(TARGET) --pkg=$(SERVERMAIN) .
	xgo --dest=$(BUILDDIR) --tags="confidant testnet" --ldflags=$(TESTLDFLAGS) --out=$(SERVERTESTBINARY)-confidant-v$(SERVERVERSION)-$(GITREV) \
	--targets=$(TARGET_CONFIDANT) --pkg=$(SERVERMAIN) .

gqlc-mining:
	xgo --dest=$(BUILDDIR) --ldflags=$(TESTLDFLAGS) --out=gqlc-miner-v$(SERVERVERSION)-$(GITREV) \
	--targets=$(TARGET) --pkg=cmd/miner/*.go .
	xgo --dest=$(BUILDDIR) --ldflags=$(TESTLDFLAGS) --out=gqlc-stratum-v$(SERVERVERSION)-$(GITREV) \
	--targets=$(TARGET) --pkg=cmd/stratum/*.go .

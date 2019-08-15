.PHONY: all clean build build-test confidant confidant-test
.PHONY: gqlc-server gqlc-server-test
.PHONY: gqlc-client
.PHONY: deps

# Check for required command tools to build or stop immediately
EXECUTABLES = git go find pwd
K := $(foreach exec,$(EXECUTABLES),\
        $(if $(shell which $(exec)),some string,$(error "No $(exec) in PATH)))

# server
SERVERVERSION ?= 1.2.4
SERVERBINARY = gqlc
SERVERTESTBINARY = gqlct
SERVERMAIN = cmd/server/main.go

# client
CLIENTVERSION ?= 1.2.4
CLIENTBINARY = gqlcc
CLIENTMAIN = cmd/client/main.go

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

CLIENTLDFLAGS="-X github.com/qlcchain/go-qlc/cmd/client/commands.Version=${CLIENTVERSION} \
	-X github.com/qlcchain/go-qlc/cmd/client/commands.GitRev=${GITREV} \
	-X github.com/qlcchain/go-qlc/cmd/client/commands.BuildTime=${BUILDTIME}" \

deps:
	go get -u golang.org/x/lint/golint
	go get -u github.com/gythialy/xgo
	go get -u github.com/git-chglog/git-chglog/cmd/git-chglog

confidant:
	CGO_ENABLED=1 CC=/opt/gcc-linaro-5.3.1-2016.05-x86_64_arm-linux-gnueabihf/bin/arm-linux-gnueabihf-gcc GOARCH=arm GOARM=7 \
	GO111MODULE=on go build -tags "confidant" -ldflags $(MAINLDFLAGS) -v -i -o $(shell pwd)/$(BUILDDIR)/$(SERVERBINARY) $(shell pwd)/$(SERVERMAIN)
	@echo "Build $(SERVERBINARY) done."
	@echo "Run \"$(shell pwd)/$(BUILDDIR)/$(SERVERBINARY)\" to start $(SERVERBINARY)."
	CGO_ENABLED=1 CC=/home/lichao/ppr_cross/gcc-linaro-5.3.1-2016.05-x86_64_arm-linux-gnueabihf/bin/arm-linux-gnueabihf-gcc GOARCH=arm GOARM=7 \
	GO111MODULE=on go build -ldflags $(CLIENTLDFLAGS) -v -i -o $(shell pwd)/$(BUILDDIR)/$(CLIENTBINARY) $(shell pwd)/$(CLIENTMAIN)
	@echo "Build $(CLIENTBINARY) done."
	@echo "Run \"$(shell pwd)/$(BUILDDIR)/$(CLIENTBINARY)\" to start $(CLIENTBINARY)."

confidant-test:
	CGO_ENABLED=1 CC=/opt/gcc-linaro-5.3.1-2016.05-x86_64_arm-linux-gnueabihf/bin/arm-linux-gnueabihf-gcc GOARCH=arm GOARM=7 \
	GO111MODULE=on go build -tags "confidant testnet" -ldflags $(MAINLDFLAGS) -v -i -o $(shell pwd)/$(BUILDDIR)/$(SERVERBINARY) $(shell pwd)/$(SERVERMAIN)
	@echo "Build $(SERVERBINARY) done."
	@echo "Run \"$(shell pwd)/$(BUILDDIR)/$(SERVERBINARY)\" to start $(SERVERBINARY)."
	CGO_ENABLED=1 CC=/home/lichao/ppr_cross/gcc-linaro-5.3.1-2016.05-x86_64_arm-linux-gnueabihf/bin/arm-linux-gnueabihf-gcc GOARCH=arm GOARM=7 \
	GO111MODULE=on go build -ldflags $(CLIENTLDFLAGS) -v -i -o $(shell pwd)/$(BUILDDIR)/$(CLIENTBINARY) $(shell pwd)/$(CLIENTMAIN)
	@echo "Build $(CLIENTBINARY) done."
	@echo "Run \"$(shell pwd)/$(BUILDDIR)/$(CLIENTBINARY)\" to start $(CLIENTBINARY)."

build:
	GO111MODULE=on go build -ldflags $(MAINLDFLAGS) -v -i -o $(shell pwd)/$(BUILDDIR)/$(SERVERBINARY) $(shell pwd)/$(SERVERMAIN)
	@echo "Build $(SERVERBINARY) done."
	@echo "Run \"$(shell pwd)/$(BUILDDIR)/$(SERVERBINARY)\" to start $(SERVERBINARY)."
	GO111MODULE=on go build -ldflags $(CLIENTLDFLAGS) -v -i -o $(shell pwd)/$(BUILDDIR)/$(CLIENTBINARY) $(shell pwd)/$(CLIENTMAIN)
	@echo "Build $(CLIENTBINARY) done."
	@echo "Run \"$(shell pwd)/$(BUILDDIR)/$(CLIENTBINARY)\" to start $(CLIENTBINARY)."

build-test:
	GO111MODULE=on go build -tags "testnet" -ldflags $(TESTLDFLAGS) -v -i -o $(shell pwd)/$(BUILDDIR)/$(SERVERBINARY) $(shell pwd)/$(SERVERMAIN)
	@echo "Build test server done."
	@echo "Run \"$(BUILDDIR)/$(SERVERBINARY)\" to start $(SERVERBINARY)."
	GO111MODULE=on go build -ldflags $(CLIENTLDFLAGS) -v -i -o $(shell pwd)/$(BUILDDIR)/$(CLIENTBINARY) $(shell pwd)/$(CLIENTMAIN)
	@echo "Build test client done."
	@echo "Run \"$(BUILDDIR)/$(CLIENTBINARY)\" to start $(CLIENTBINARY)."

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

gqlc-client:
	xgo --dest=$(BUILDDIR) --ldflags=$(CLIENTLDFLAGS) --out=$(CLIENTBINARY)-v$(CLIENTVERSION)-$(GITREV) \
	--targets="windows-6.0/amd64,darwin-10.10/amd64,linux/amd64" \
	--pkg=$(CLIENTMAIN) .

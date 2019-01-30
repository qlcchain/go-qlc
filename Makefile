
.PHONY: all clean
.PHONY: gqlc_linux  gqlc-linux-amd64 gqlc-darwin-amd64
.PHONY: gqlc-darwin gqlc-darwin-amd64
.PHONY: gqlc-windows gqlc-windows-386 gqlc-windows-amd64

# Check for required command tools to build or stop immediately
EXECUTABLES = git go find pwd
K := $(foreach exec,$(EXECUTABLES),\
        $(if $(shell which $(exec)),some string,$(error "No $(exec) in PATH)))

GO ?= latest

BINARY = gqlc
SERVERMAIN = $(shell pwd)/cmd/main.go
BUILDDIR = $(shell pwd)/build
VERSION = 0.0.5
GITREV = $(shell git rev-parse --short HEAD)
BUILDTIME = $(shell date +'%Y-%m-%d_%T')
LDFLAGS=-ldflags "-X main.version=${VERSION} -X main.sha1ver=${GITREV} -X main.buildTime=${BUILDTIME}"

build:
	go build ${LDFLAGS} -v -i -o $(BUILDDIR)/$(BINARY) $(SERVERMAIN)
	@echo "Build server done."
	@echo "Run \"$(BUILDDIR)/$(BINARY)\" to start gqlc."

all: gqlc-windows gqlc-darwin gqlc-linux

clean:
	rm -rf $(BUILDDIR)/

gqlc-linux: gqlc-linux-amd64
	@echo "Linux cross compilation done:"

gqlc-linux-amd64:
	env GOOS=linux GOARCH=amd64 go build ${LDFLAGS} -i -o $(BUILDDIR)/$(BINARY)-linux-amd64-v$(VERSION)-$(GITREV) $(SERVERMAIN)
	@echo "Build linux server done."
	@ls -ld $(BUILDDIR)/$(BINARY)-linux-amd64-v$(VERSION)-$(GITREV)

gqlc-darwin:
	env GOOS=darwin GOARCH=amd64 go build ${LDFLAGS} -i -o $(BUILDDIR)/$(BINARY)-darwin-amd64-v$(VERSION)-$(GITREV) $(SERVERMAIN)
	@echo "Build darwin server done."
	@ls -ld $(BUILDDIR)/$(BINARY)-darwin-amd64-v$(VERSION)-$(GITREV)

gqlc-windows: gqlc-windows-amd64 gqlc-windows-386
	@echo "Windows cross compilation done:"
	@ls -ld $(BUILDDIR)/$(BINARY)-windows-*

gqlc-windows-386:
	env GOOS=windows GOARCH=386 go build ${LDFLAGS} -i -o $(BUILDDIR)/$(BINARY)-windows-386-v$(VERSION)-$(GITREV).exe $(SERVERMAIN)
	@echo "Build windows x86 server done."
	@ls -ld $(BUILDDIR)/$(BINARY)-windows-386-v$(VERSION)-$(GITREV).exe

gqlc-windows-amd64:
	env GOOS=windows GOARCH=amd64 go build ${LDFLAGS} -i -o $(BUILDDIR)/$(BINARY)-windows-amd64-v$(VERSION)-$(GITREV).exe $(SERVERMAIN)
	@echo "Build windows server done."
	@ls -ld $(BUILDDIR)/$(BINARY)-windows-amd64-v$(VERSION)-$(GITREV).exe

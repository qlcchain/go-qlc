.PHONY: all clean
.PHONY: gqlc_linux  gqlc-linux-amd64 gqlc-darwin-amd64
.PHONY: gqlc-darwin gqlc-darwin-amd64
.PHONY: gqlc-windows gqlc-windows-386 gqlc-windows-amd64
.PHONY: gqlcc_linux  gqlcc-linux-amd64 gqlcc-darwin-amd64
.PHONY: gqlcc-darwin gqlcc-darwin-amd64
.PHONY: gqlcc-windows gqlcc-windows-386 gqlcc-windows-amd64

# Check for required command tools to build or stop immediately
EXECUTABLES = git go find pwd
K := $(foreach exec,$(EXECUTABLES),\
        $(if $(shell which $(exec)),some string,$(error "No $(exec) in PATH)))

GO ?= latest

SERVERBINARY = gqlc
SERVERMAIN = $(shell pwd)/cmd/server/main.go
CLIENTBINARY = gqlcc
CLIENTMAIN = $(shell pwd)/cmd/client/main.go

BUILDDIR = $(shell pwd)/build
VERSION = $(shell cat buildversion)
GITREV = $(shell git rev-parse --short HEAD)
BUILDTIME = $(shell date +'%Y-%m-%d_%T')
LDFLAGS=-ldflags "-X main.version=${VERSION} -X main.sha1ver=${GITREV} -X main.buildTime=${BUILDTIME}"


build:
	go build ${LDFLAGS} -v -i -o $(BUILDDIR)/$(SERVERBINARY) $(SERVERMAIN)
	@echo "Build server done."
	@echo "Run \"$(BUILDDIR)/$(SERVERBINARY)\" to start gqlc."
	go build ${LDFLAGS} -v -i -o $(BUILDDIR)/$(CLIENTBINARY) $(CLIENTMAIN)
	@echo "Build client done."
	@echo "Run \"$(BUILDDIR)/$(CLIENTBINARY)\" to start gqlcc."

all: gqlc-windows gqlc-darwin gqlc-linux gqlcc-windows gqlcc-darwin gqlcc-linux

clean:
	rm -rf $(BUILDDIR)/

gqlc-linux: gqlc-linux-amd64
	@echo "Linux cross compilation done:"

gqlc-linux-amd64:
	env GOOS=linux GOARCH=amd64 go build ${LDFLAGS} -i -o $(BUILDDIR)/$(SERVERBINARY)-linux-amd64-v$(VERSION)-$(GITREV) $(SERVERMAIN)
	@echo "Build linux server done."
	@ls -ld $(BUILDDIR)/$(SERVERBINARY)-linux-amd64-v$(VERSION)-$(GITREV)

gqlc-darwin:
	env GOOS=darwin GOARCH=amd64 go build ${LDFLAGS} -i -o $(BUILDDIR)/$(SERVERBINARY)-darwin-amd64-v$(VERSION)-$(GITREV) $(SERVERMAIN)
	@echo "Build darwin server done."
	@ls -ld $(BUILDDIR)/$(SERVERBINARY)-darwin-amd64-v$(VERSION)-$(GITREV)

gqlc-windows: gqlc-windows-amd64 gqlc-windows-386
	@echo "Windows cross compilation done:"
	@ls -ld $(BUILDDIR)/$(SERVERBINARY)-windows-*

gqlc-windows-386:
	env GOOS=windows GOARCH=386 go build ${LDFLAGS} -i -o $(BUILDDIR)/$(SERVERBINARY)-windows-386-v$(VERSION)-$(GITREV).exe $(SERVERMAIN)
	@echo "Build windows x86 server done."
	@ls -ld $(BUILDDIR)/$(SERVERBINARY)-windows-386-v$(VERSION)-$(GITREV).exe

gqlc-windows-amd64:
	env GOOS=windows GOARCH=amd64 go build ${LDFLAGS} -i -o $(BUILDDIR)/$(SERVERBINARY)-windows-amd64-v$(VERSION)-$(GITREV).exe $(SERVERMAIN)
	@echo "Build windows server done."
	@ls -ld $(BUILDDIR)/$(SERVERBINARY)-windows-amd64-v$(VERSION)-$(GITREV).exe



gqlcc-linux: gqlcc-linux-amd64
	@echo "Linux cross compilation done:"

gqlcc-linux-amd64:
	env GOOS=linux GOARCH=amd64 go build ${LDFLAGS} -i -o $(BUILDDIR)/$(CLIENTBINARY)-linux-amd64-v$(VERSION)-$(GITREV) $(CLIENTMAIN)
	@echo "Build linux client done."
	@ls -ld $(BUILDDIR)/$(CLIENTBINARY)-linux-amd64-v$(VERSION)-$(GITREV)

gqlcc-darwin:
	env GOOS=darwin GOARCH=amd64 go build ${LDFLAGS} -i -o $(BUILDDIR)/$(CLIENTBINARY)-darwin-amd64-v$(VERSION)-$(GITREV) $(CLIENTMAIN)
	@echo "Build darwin client done."
	@ls -ld $(BUILDDIR)/$(CLIENTBINARY)-darwin-amd64-v$(VERSION)-$(GITREV)

gqlcc-windows: gqlcc-windows-amd64 gqlcc-windows-386
	@echo "Windows cross compilation done:"
	@ls -ld $(BUILDDIR)/$(CLIENTBINARY)-windows-*

gqlcc-windows-386:
	env GOOS=windows GOARCH=386 go build ${LDFLAGS} -i -o $(BUILDDIR)/$(CLIENTBINARY)-windows-386-v$(VERSION)-$(GITREV).exe $(CLIENTMAIN)
	@echo "Build windows x86 client done."
	@ls -ld $(BUILDDIR)/$(CLIENTBINARY)-windows-386-v$(VERSION)-$(GITREV).exe

gqlcc-windows-amd64:
	env GOOS=windows GOARCH=amd64 go build ${LDFLAGS} -i -o $(BUILDDIR)/$(CLIENTBINARY)-windows-amd64-v$(VERSION)-$(GITREV).exe $(CLIENTMAIN)
	@echo "Build windows client done."
	@ls -ld $(BUILDDIR)/$(CLIENTBINARY)-windows-amd64-v$(VERSION)-$(GITREV).exe

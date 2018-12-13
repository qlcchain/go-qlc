#
# Copyright (c) 2018 QLC Chain Team
# This software is released under the MIT License.
# https://opensource.org/licenses/MIT
#

# Check for required command tools to build or stop immediately
EXECUTABLES = git go find pwd
K := $(foreach exec,$(EXECUTABLES),\
        $(if $(shell which $(exec)),some string,$(error "No $(exec) in PATH)))

GO ?= latest

MAIN = $(shell pwd)/cmd/main.go
BUILDDIR = $(shell pwd)/build
GOBIN = $(BUILDDIR)
BINARY = gqlc
VERSION = 0.0.1
GITREV = $(shell git rev-parse --short HEAD)
BUILDTIME = $(shell date +'%Y-%m-%d_%T')

PLATFORMS = darwin linux windows
ARCHITECTURES = 386 amd64

# Setup linker flags option for build that interoperate with variable names in src code
LDFLAGS=-ldflags "-X main.version=${VERSION} -X main.sha1ver=${GITREV} -X main.buildTime=${BUILDTIME}"

default: build
all: clean build_all

build:
	go build ${LDFLAGS} -i -o $(GOBIN)/${BINARY} $(MAIN)
	@echo "Build $(BINARY) done.\nRun \"$(GOBIN)/$(BINARY)\" to start it."

build_all:
	$(foreach GOOS, $(PLATFORMS),\
	$(foreach GOARCH, $(ARCHITECTURES), $(shell export GOOS=$(GOOS); export GOARCH=$(GOARCH); go build ${LDFLAGS} -i -o $(GOBIN)/$(GOOS)/$(BINARY)-$(GOOS)-$(GOARCH) $(MAIN))))

clean:
	rm -rf $(BUILDDIR)/
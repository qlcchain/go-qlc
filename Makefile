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

MAIN = $(shell pwd)/cmd/config.go
BUILDDIR = $(shell pwd)/build
GOBIN = $(BUILDDIR)
BINARY = go-qlc
VERSION =0.0.1
GITREV = $(shell git rev-parse HEAD)

PLATFORMS = darwin linux windows
ARCHITECTURES = 386 amd64

# Setup linker flags option for build that interoperate with variable names in src code
LDFLAGS=-ldflags "-X main.Version=${VERSION} -X main.Build=${GITREV}"

default: build
all: clean build_all

build:
	go build ${LDFLAGS} -i -o $(GOBIN)/${BINARY} $(MAIN)
	@echo "Build go-qlc done.\nRun \"$(GOBIN)/go-qlc\" to start go-qlc."

build_all:
	$(foreach GOOS, $(PLATFORMS),\
	$(foreach GOARCH, $(ARCHITECTURES), $(shell export GOOS=$(GOOS); export GOARCH=$(GOARCH); go build ${LDFLAGS} -i -o $(GOBIN)/$(GOOS)/$(BINARY)-$(GOOS)-$(GOARCH) $(MAIN))))

clean:
	rm -rf $(BUILDDIR)/
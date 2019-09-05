#!/usr/bin/env bash

set -e

if command -v golangci-lint ; then
    echo "golangci-lint already exist"
else
    go get -u github.com/golangci/golangci-lint/cmd/golangci-lint
fi

golangci-lint run ./...

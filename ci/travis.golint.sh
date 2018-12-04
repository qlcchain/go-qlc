#!/usr/bin/env bash

set -e

which golangci-lint
if [[ $? -eq 0 ]]; then
    echo "golangci-lint already exist"
else
    go get -u github.com/golangci/golangci-lint/cmd/golangci-lint
fi

golangci-lint run ./...
#!/usr/bin/env bash

set -e
echo "" > coverage.txt

for d in $(go list ./... | grep -v vendor | grep -v edwards25519); do
#    go test -race -coverprofile=profile.out -covermode=atomic "$d"
    go test -coverprofile=profile.out "$d"
    if [[ -f profile.out ]]; then
        cat profile.out >> coverage.txt
        rm profile.out
    fi
done
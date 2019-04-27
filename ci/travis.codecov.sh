#!/usr/bin/env bash

set -e
echo "" > coverage.txt

tags=(mainnet testnet)
integrate=""

if [[ -n "$TRAVIS_TAG" ]]; then
    integrate="integrate"
fi

for t in "${tags[@]}"; do
    echo "start test case for $t"
    for d in $(go list ./... | grep -v vendor | grep -v edwards25519); do
    #    go test -race -coverprofile=profile.out -covermode=atomic "$d"
        go test -tags "$t $integrate" -coverprofile=profile.out "$d"
        if [[ -f profile.out ]]; then
            cat profile.out >> coverage.txt
            rm profile.out
        fi
    done
done
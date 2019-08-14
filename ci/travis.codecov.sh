#!/usr/bin/env bash

set -e
echo "" > coverage.txt

tags=(mainnet testnet)

for t in "${tags[@]}"; do
  echo "run test cases by $t"
  for d in $(go list ./... | grep -v vendor | grep -v edwards25519); do
#    go test -race -coverprofile=profile.out -covermode=atomic "$d"
    go test -tags "$t" -coverprofile=profile.out "$d"
    if [[ -f profile.out ]]; then
      cat profile.out >> coverage.txt
      rm profile.out
    fi
  done
done

#if [[ -n "$TRAVIS_TAG" ]]; then
#    echo 'build tag for release, run all test cases'
#    for d in $(go list ./... | grep -v vendor | grep -v edwards25519); do
#        go test -tags "testnet integrate" -coverprofile=profile.out "$d"
#        if [[ -f profile.out ]]; then
#            cat profile.out >> coverage.txt
#            rm profile.out
#        fi
#    done
#fi

#!/usr/bin/env bash

set -e

REPO_ROOT=`git rev-parse --show-toplevel`
cd $REPO_ROOT
DIRS="common consensus p2p rpc wallet chain cmd config ledger vm"
tags=(mainnet testnet)

for subdir in $DIRS; do
    pushd $subdir
    for t in "${tags[@]}"; do
        go vet -tags "$t"
    done
    popd
done

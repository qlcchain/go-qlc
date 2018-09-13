#!/usr/bin/env bash

set -e

REPO_ROOT=`git rev-parse --show-toplevel`
cd $REPO_ROOT
DIRS="common consensus p2p rpc wallet chain cmd config ledger qlcpb"

for subdir in $DIRS; do
  pushd $subdir
  go vet
  popd
done

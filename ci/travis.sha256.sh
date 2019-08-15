#!/usr/bin/env bash

set -e

if command -v sha256sum; then
    if [[ -d "build" ]]; then
        files=$(find build -type f)
        for f in ${files}; do
            fn="${f%.exe}.tar.gz"
#            echo "${f} ==> ${fn}"
            tar -czf "${fn}" "${f}"
            sha256sum "${fn}"
        done
    fi
else
    echo 'can not find sha256sum.'
fi

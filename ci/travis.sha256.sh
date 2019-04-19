#!/usr/bin/env bash

set -e

which sha256sum

if [[ $? -eq 0 ]]; then
    if [[ -d "build" ]]; then
        files=$(find build -type f)
        for f in ${files}; do
            echo $(sha256sum ${f})
        done
    fi
else
    echo 'can not find sha256sum.'
fi


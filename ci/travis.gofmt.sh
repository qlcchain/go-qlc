#!/usr/bin/env bash

goimports -w -local github.com/qlcchain/go-qlc $(find . -type f -name '*.go' -not -path "*/mocks/*" -not -path "*/pb/*" -not -path "*/proto/*")

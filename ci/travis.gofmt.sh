#!/usr/bin/env bash

if [[ -n "$(gofmt -l .| grep -v ^vendor/)" ]]; then
  echo "Go code is not formatted:"
  gofmt -d .
  exit 1
fi

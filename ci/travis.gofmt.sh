#!/usr/bin/env bash

if [[ -n "$(goimports -l .| grep -v ^vendor/)" ]]; then
  echo "Go code is not formatted:"
  goimports -d .
  exit 1
fi
#!/usr/bin/env bash

set -e

function build() {
  # check
  GOOS=$1
  GOARCH=$2
  if [ -z "$GOOS" ] || [ -z "$GOARCH" ]; then
    echo "GOOS or GOARCH is empty"
    exit 1
  fi

  # output
  OUTPUT="rocket-proxy_${GOOS}_${GOARCH}"
  if [ "$GOOS" = "windows" ]; then
    OUTPUT="$OUTPUT.exe"
  fi
  rm -rf ./build && mkdir -p ./build

  # build
  echo "building, os: $GOOS, arch: $GOARCH, output: $OUTPUT"
  GOOS=$1 GOARCH=$2 go build -v -a -ldflags '-s -w' \
   -gcflags="all=-trimpath=${PWD}" \
   -asmflags="all=-trimpath=${PWD}" \
   -o ./build/$OUTPUT \
   ./cmd/*.go
  # compress
  if [ -x "$(command -v upx)" ]; then
    upx ./build/$OUTPUT
  fi
}

# build linux
build linux amd64

#!/usr/bin/env bash

set -e

BUILD_VERSION=$(date +%Y.%m.%d)
TRIM_PATH=$(pwd)

function build() {
  # check
  GOOS=$1
  GOARCH=$2
  if [ -z "$GOOS" ] || [ -z "$GOARCH" ]; then
    echo "GOOS or GOARCH is empty"
    exit 1
  fi

  # output
  OUTPUT="fluxproxy_${GOOS}_${GOARCH}"
  if [ "$GOOS" = "windows" ]; then
    OUTPUT="$OUTPUT.exe"
  fi
  rm -rf ./build && mkdir -p ./build

  # build
  echo "BUILD: version: $BUILD_VERSION, os: $GOOS, arch: $GOARCH, output: $OUTPUT"
  echo "BUILD: trimpath: $TRIM_PATH"
  GOOS=$1 GOARCH=$2 go build -v -a -ldflags "-s -w -X 'main.BuildVersion=$BUILD_VERSION'" \
    -gcflags="all=-trimpath=$TRIM_PATH" \
    -asmflags="all=-trimpath=$TRIM_PATH" \
    -o ./build/$OUTPUT \
    ./cmd/*.go
  # compress
  if [ -x "$(command -v upx)" ]; then
    upx ./build/$OUTPUT
  fi
}

# build linux
build linux amd64

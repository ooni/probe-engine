#!/bin/sh
set -e
DISABLE_QUIC=""
if [ ! -z "$(go version | grep go1.15)" ]; then DISABLE_QUIC="DISABLE_QUIC"; fi

set -ex
case $1 in
  darwin)
    export GOOS=darwin GOARCH=amd64
    go build -o ./CLI/darwin/amd64 -tags $DISABLE_QUIC -ldflags="-s -w" ./cmd/miniooni;;
  linux)
    export GOOS=linux GOARCH=amd64
    go build -o ./CLI/linux/amd64 -tags $DISABLE_QUIC,netgo -ldflags='-s -w -extldflags "-static"' ./cmd/miniooni;;
  windows)
    export GOOS=windows GOARCH=amd64
    go build -o ./CLI/windows/amd64 -tags $DISABLE_QUIC -ldflags="-s -w" ./cmd/miniooni;;
  *)
    echo "usage: $0 darwin|linux|windows" 1>&2
    exit 1
esac

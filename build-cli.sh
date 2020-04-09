#!/bin/sh
set -ex
case $1 in
  _linux)
    apk add gcc musl-dev
    go build -tags "netgo nomk" -ldflags='-s -w -extldflags "-static"' \
      -o ./CLI/linux/amd64/miniooni ./cmd/miniooni;;
  darwin)
    go build -tags nomk -o ./CLI/darwin/amd64 ./cmd/miniooni;;
  linux)
    docker run -it -v`pwd`:/ooni -w/ooni golang:alpine ./build-cli.sh _linux;;
  windows)
    export CC=x86_64-w64-mingw32-gcc
    go build -tags -nomk -o ./CLI/windows/amd64 ./cmd/miniooni;;
  *)
    echo "usage: $0 darwin|linux|windows" 1>&2
    exit 1
esac

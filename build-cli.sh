#!/bin/sh
set -ex
case $1 in
  darwin)
    export GOOS=darwin GOARCH=amd64
    go build -o ./CLI/darwin/amd64 -ldflags="-s -w" ./cmd/miniooni;;
  linux)
    export GOOS=linux GOARCH=amd64
    go build -o ./CLI/linux/amd64 -tags netgo -ldflags='-s -w -extldflags "-static"' ./cmd/miniooni;;
  windows)
    export GOOS=windows GOARCH=amd64
    go build -o ./CLI/windows/amd64 -ldflags="-s -w" ./cmd/miniooni;;
  rpi3)
    export GOOS=linux GOARCH=arm GOARM=7
    go build -o ./CLI/linux/armv7 -tags netgo -ldflags='-s -w -extldflags "-static"' ./cmd/miniooni;;
  rpi4)
    export GOOS=linux GOARCH=arm64
    go build -o ./CLI/linux/arm64 -tags netgo -ldflags='-s -w -extldflags "-static"' ./cmd/miniooni;;
  *)
    echo "usage: $0 darwin|linux|windows|rpi3|rpi4" 1>&2
    exit 1
esac

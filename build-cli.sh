#!/bin/sh
set -ex
case $1 in
  darwin)
    export GOOS=darwin GOARCH=amd64
    go build -o ./CLI/darwin/amd64 -ldflags="-s -w" ./cmd/miniooni;;
    echo ""
    echo "Binary ready at ./CLI/darwin/amd64/miniooni"
  linux)
    export GOOS=linux GOARCH=amd64
    go build -o ./CLI/linux/amd64 -tags netgo -ldflags='-s -w -extldflags "-static"' ./cmd/miniooni;;
    echo ""
    echo "Binary ready at ./CLI/linux/amd64/miniooni"
  windows)
    export GOOS=windows GOARCH=amd64
    go build -o ./CLI/windows/amd64 -ldflags="-s -w" ./cmd/miniooni;;
    echo ""
    echo "Binary ready at ./CLI/windows/amd64/miniooni.exe"
  *)
    echo "usage: $0 darwin|linux|windows" 1>&2
    exit 1
esac

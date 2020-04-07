#!/bin/bash
set -ex
export CGO_ENABLED=1
case $1 in
  linux)
    export GOOS=linux
    export GOARCH=amd64
    dll=so
    ;;
  macos)
    export GOOS=darwin
    export GOARCH=amd64
    dll=so
    ;;
  windows)
    export GOOS=windows
    export GOARCH=amd64
    export CC=x86_64-w64-mingw32-gcc
    dll=dll
    ;;
  *)
    echo "Usage: $0 linux|macos|windows" 1>&2
    exit 1
esac
go build -x -buildmode c-shared -tags nomk -ldflags="-s -w" -o libminiooni.$dll ./lib/libminiooni/...
rm libminiooni.h
install -d ./DLL/$1/
mv libminiooni.$dll ./DLL/$1/

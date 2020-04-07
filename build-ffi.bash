#!/bin/bash
set -e
case $1 in
  darwin|linux)
    so=so;;
  windows)
    export CC=${CC:-x86_64-w64-mingw32-gcc}
    so=dll;;
  *)
    echo "usage: $0 darwin|linux|windows" 1>&2
    exit 1
esac
export CGO_ENABLED=1 GOOS=$1 GOARCH=amd64
out=./FFI/$GOOS/$GOARCH
rm -rf $out
install -d $out
go build -x -buildmode c-shared -tags nomk -ldflags='-s -w' \
    -o $out/libooniffi.$so ./libooniffi/...
rm $out/libooniffi.h
cp ./libooniffi/ooniffi.h $out/

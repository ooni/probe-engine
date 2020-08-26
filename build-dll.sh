#!/bin/sh
set -e
case $1 in
  darwin)
    prefix=lib
    ext=so
    export GOOS=darwin
    ;;
  linux)
    prefix=lib
    ext=so
    export GOOS=linux
    ;;
  windows)
    export CC=x86_64-w64-mingw32-gcc
    ext=dll
    export GOOS=windows
    ;;
  *)
    echo "usage: $0 darwin|linux|windows" 1>&2
    exit 1
esac
export GOARCH=amd64
export CGO_ENABLED=1
outfile="./DLL/$GOOS/$GOARCH/${prefix}ooniffi.$ext"
go build -o $outfile -buildmode=c-shared -ldflags="-s -w" ./libooniffi/...
chmod +x $outfile
rm ./DLL/$GOOS/$GOARCH/${prefix}ooniffi.h
cp ./libooniffi/ooniffi.h ./DLL/$GOOS/$GOARCH/

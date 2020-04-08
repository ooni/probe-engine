#!/bin/bash
set -ex
case $1 in
  darwin|linux)
    dll=libooniffi.so;;
  windows)
    export CC=${CC:-x86_64-w64-mingw32-gcc}
    dll=ooniffi.dll;;
  *)
    echo "usage: $0 darwin|linux|windows" 1>&2
    exit 1
esac
export CGO_ENABLED=1 GOOS=$1 GOARCH=amd64
out=./FFI/$GOOS/$GOARCH
go build -x -buildmode c-shared -tags nomk -ldflags='-s -w' \
    -o $out/$dll ./libooniffi/...
rm -f $out/libooniffi.h
cp ./libooniffi/ooniffi.h $out/
case $1 in
  windows)
    DLLTOOL=${DLLTOOL:-x86_64-w64-mingw32-dlltool}
    $DLLTOOL -d $out/ooniffi.def -l $out/ooniffi.lib $out/ooniffi.dll
    ;;
esac

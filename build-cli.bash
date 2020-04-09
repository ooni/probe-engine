#!/bin/bash
set -ex
case $1 in
  darwin|linux)
    dll=libooniffi.so
    exe=miniooni
    ;;
  windows)
    export CC=${CC:-x86_64-w64-mingw32-gcc}
    dll=ooniffi.dll
    exe=miniooni.exe
    ;;
  *)
    echo "usage: $0 darwin|linux|windows" 1>&2
    exit 1
esac
export CGO_ENABLED=1 GOOS=$1 GOARCH=amd64
out=./CLI/$GOOS/$GOARCH
function build() {
  go build -tags nomk -ldflags="-s -w" $@
}
build -o $out/$exe ./cmd/miniooni
build -buildmode c-shared -o $out/$dll ./libooniffi/...
rm -f $out/libooniffi.h
case $1 in
  linux)
    chmod +x $out/$dll
    ;;
  windows)
    DLLTOOL=${DLLTOOL:-x86_64-w64-mingw32-dlltool}
    $DLLTOOL -d $out/ooniffi.def -l $out/ooniffi.lib $out/ooniffi.dll
    ;;
esac

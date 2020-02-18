#!/bin/bash
set -e

topdir=$(cd $(dirname $0) && pwd -P)
cd $topdir

function library() {
    if [ $# -ne 2 ]; then
        echo "usage: library <goos> <goarch>" 1>&2
        exit 1
    fi
    local goos=$1
    local goarch=$2
    local distbasedir=./dist/$goos/$goarch
    rm -rf $distbasedir
    local libdir=$distbasedir/lib
    mkdir -p $libdir
    local includedir=$distbasedir/include/ooni
    mkdir -p $includedir
    local library=$libdir/libooni.so
    local buildmode=c-shared
    echo "compile $goos/$goarch"
    set -x
    cp ./libooni/ffi.h $includedir
    CGO_ENABLED=1 GOOS=$goos GOARCH=$goarch go build -v -o $library \
	    -tags nomk -buildmode=$buildmode ./libooni/...
    set +x
}

library linux amd64
library windows amd64

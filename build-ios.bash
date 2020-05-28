#!/bin/bash
set -e
topdir=$(cd $(dirname $0) && pwd -P)
set -x
export GOPATH=$topdir/MOBILE/gopath
export PATH=$GOPATH/bin:$PATH
export GO111MODULE=off
output=MOBILE/dist/oonimkall.framework
go get -u golang.org/x/mobile/cmd/gomobile
gomobile init
export GO111MODULE=on
gomobile bind -target=ios -o $output -tags nomk -ldflags="-s -w" ./oonimkall

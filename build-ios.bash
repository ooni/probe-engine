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
# See https://github.com/ooni/probe-engine/issues/668
for header in $output/Headers/*.objc.h; do
  cat $header | sed 's|^@import Foundation;|#import <Foundation/Foundation.h>|g' > $header.new
  mv $header.new $header
done

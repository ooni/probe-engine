#!/bin/bash
set -e
topdir=$(cd $(dirname $0) && pwd -P)
set -x
export PATH=$(go env GOPATH)/bin:$PATH
export GO111MODULE=off
go get -u golang.org/x/mobile/cmd/gomobile
gomobile init
export GO111MODULE=on
output=MOBILE/ios/oonimkall.framework
gomobile bind -target=ios -o $output -ldflags="-s -w" ./oonimkall
# See https://github.com/ooni/probe-engine/issues/668
for header in $output/Headers/*.objc.h; do
  cat $header | sed 's|^@import Foundation;|#import <Foundation/Foundation.h>|g' > $header.new
  mv $header.new $header
done

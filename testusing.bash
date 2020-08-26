#!/bin/bash

#
# This script simulates a user creating a new project that depends
# on github.com/ooni/probe-engine@GITHUB_SHA.
#

set -ex
mkdir -p example.org/x
cd example.org/x
go mod init example.org/x
echo > main.go << EOF
package main

import "github.com/ooni/probe-engine/libminiooni

func main() {
    libminiooni.Main()
}
EOF
go get -v github.com/ooni/probe-engine@$GITHUB_SHA
go build -v .
./x -OTunnel=psiphon -ni https://www.example.com urlgetter

#!/bin/sh
set -ex
apk add --no-progress git go
go test -tags ooni,shaping ./...

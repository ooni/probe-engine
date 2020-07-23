#!/bin/sh
set -ex
DOCKER=${DOCKER:-docker}
$DOCKER build -t jafar-qa ./cmd/jafar/
$DOCKER run --privileged -v`pwd`:/jafar -w/jafar jafar-qa "$@"

#!/bin/bash
set -e
if [ -z "$ANDROID_HOME" -o "$1" = "--help" ]; then
    echo ""
    echo "usage: $0"
    echo ""
    echo "Please set ANDROID_HOME. We assume you have installed"
    echo "the Android SDK. You can do that on macOS using:"
    echo ""
    echo "    brew cask install android-sdk"
    echo ""
    echo "Then make sure you install the required packages:"
    echo ""
    echo "sdkmanager --install 'build-tools;29.0.3' ndk-bundle"
    echo ""
    echo "or, if you already installed, that you're up to date:"
    echo ""
    echo "sdkmanager --update"
    echo ""
    echo "Once you have done that, please export ANDROID_HOME to"
    echo "point to /usr/local/Caskroom/android-sdk/<version>."
    echo ""
    exit 1
fi
topdir=$(cd $(dirname $0) && pwd -P)
set -x
export GOPATH=$topdir/MOBILE/gopath
export PATH=$GOPATH/bin:$PATH
export GO111MODULE=off
GITHUB_SHA=${GITHUB_SHA:-git rev-parse HEAD}
GITHUB_RUN_ID=${GITHUB_RUN_ID:0}
timestamp=$(date -u +%Y%m%dT%H%M%SZ)
version="$(echo $GITHUB_SHA|cut -c-10)-$GITHUB_RUN_ID-$timestamp"
output=MOBILE/dist/oonimkall-$version.aar
go get -u golang.org/x/mobile/cmd/gomobile
gomobile init
export GO111MODULE=on
gomobile bind -target=android -o $output -tags nomk -ldflags="-s -w" ./oonimkall

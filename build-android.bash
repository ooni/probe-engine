#!/bin/bash
set -e
test -z $ANDROID_HOME && {
    echo ""
    echo "$0: fatal: please set ANDROID_HOME."
    echo ""
    echo "We assume you have installed the Android SDK. You can do"
    echo "that on macOS by running this command:"
    echo ""
    echo "    brew cask install android-sdk"
    echo ""
    echo "Once you have done that, please export ANDROID_HOME to"
    echo "point to /usr/local/Caskroom/android-sdk/<version>."
    echo ""
    exit 1
}
topdir=$(cd $(dirname $0) && pwd -P)
set -x
$ANDROID_HOME/tools/bin/sdkmanager --install 'build-tools;29.0.3' ndk-bundle
$ANDROID_HOME/tools/bin/sdkmanager --update --verbose
export GOPATH=$topdir/MOBILE/gopath
export PATH=$GOPATH/bin:$PATH
export GO111MODULE=off
version="$(git describe --tags --dirty)-$(date -u +%Y%m%dT%H%M%SZ)"
output=MOBILE/dist/oonimkall-$version.aar
go get -u -v golang.org/x/mobile/cmd/gomobile
gomobile init
export GO111MODULE=on
gomobile bind -target=android -o $output -v -tags nomk -ldflags="-s -w" ./oonimkall

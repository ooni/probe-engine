#!/bin/bash
set -e

topdir=$(cd $(dirname $0) && pwd -P)
cd $topdir

function verbose() {
  echo "+ $@"
  "$@"
}

function build_libooni() {
  if [ -z $GOOS -o -z $GOARCH ]; then
    echo "fatal: build_libooni requires GOOS and GOARCH to be set" 1>&2
    exit 1
  fi
  if [ $# -ne 3 ]; then
    echo "usage: build_libooni <destdir> <buildmode> <libooni_name>" 1>&2
    exit 1
  fi
  local destdir=$1
  local includedir=$destdir/include/ooni
  verbose rm -rf $includedir
  verbose install -d $includedir
  local libdir=$destdir/lib
  verbose rm -rf $libdir
  local buildmode=$2
  local libooni_name=$3
  local output=$libdir/$libooni_name
  verbose cp ./libooni/libooni/ffi.h $includedir
  verbose go build -v -tags nomk -ldflags="-s -w" -buildmode=$buildmode -o $output ./libooni/libooni
  verbose rm -f $libdir/libooni.h
}

function missing_ooni_android_toolchain_error() {
  cat << EOF
You're missing the OONI_ANDROID_TOOLCHAIN environment variable. This variable
must point to the precompiled clang binary in your Android toolchain.

On macOS, follow these steps to install:

1. brew install android-sdk
2. brew tap adoptopenjdk/openjdk
3. brew cask install adoptopenjdk8
4. sdkmanager --install ndk-bundle

Once you have installed, run:

1. sdkmanager --update

To properly set OONI_ANDROID_TOOLCHAIN, run:

export OONI_ANDROID_TOOLCHAIN=\$(dirname \$(dirname \$(find /usr/local/Caskroom/android-sdk/ -type f -name clang)))
EOF
}

if [ "$1" = "android" ]; then
  if [ -z $OONI_ANDROID_TOOLCHAIN ]; then
    missing_ooni_android_toolchain_error
    exit 1
  fi
  verbose export CGO_ENABLED=1
  verbose export CC=$OONI_ANDROID_TOOLCHAIN/bin/aarch64-linux-android21-clang
  verbose export GOOS=android
  verbose export GOARCH=arm64
  build_libooni dist/$GOOS/$GOARCH c-shared libooni.so
elif [ "$1" = "linux" ]; then
  verbose export GOOS=linux
  verbose export GOARCH=amd64
  build_libooni dist/$GOOS/$GOARCH c-shared libooni.so
elif [ "$1" = "macos" ]; then
  verbose export GOOS=darwin
  verbose export GOARCH=amd64
  build_libooni dist/macos/$GOARCH c-shared libooni.dylib
elif [ "$1" = "windows" ]; then
  verbose export CC=x86_64-w64-mingw32-gcc
  verbose export CGO_ENABLED=1
  verbose export GOOS=windows
  verbose export GOARCH=amd64
  build_libooni dist/$GOOS/$GOARCH c-shared libooni.dll
else
  echo "usage: $0 android|linux|macos|windows" 1>&2
  exit 1
fi

[![GoDoc](https://godoc.org/github.com/ooni/probe-engine?status.svg)](https://godoc.org/github.com/ooni/probe-engine) [![Build Status](https://travis-ci.org/ooni/probe-engine.svg?branch=master)](https://travis-ci.org/ooni/probe-engine) [![Coverage Status](https://coveralls.io/repos/github/ooni/probe-engine/badge.svg?branch=master)](https://coveralls.io/github/ooni/probe-engine?branch=master) [![Go Report Card](https://goreportcard.com/badge/github.com/ooni/probe-engine)](https://goreportcard.com/report/github.com/ooni/probe-engine)

# OONI probe measurement engine

This repository contains OONI probe's [measurement engine](
https://github.com/ooni/spec/tree/master/probe#engine). That is, the
piece of software that implements OONI nettests.

## API

You can [browse ooni/probe-engine's API](
https://godoc.org/github.com/ooni/probe-engine?status.svg)
at [godoc.org](https://godoc.org/). We don't provide any API
stability guarantees and, as such, we will never release v1.0.0
of this software.

This repository also allows to build [miniooni](cmd/miniooni), a
small command line client useful to test the functionality in here
without integrating with OONI probe. You can browse [the manual
of this tool](
https://godoc.org/github.com/ooni/probe-engine/cmd/miniooni)
at [godoc.org](https://godoc.org/). We may change the command line
API of miniooni without notice.

## Integrating ooni/probe-engine

This software uses [Go modules](https://github.com/golang/go/wiki/Modules)
and therefore requires Go v1.11+. We also depend on [Measurement Kit](
https://github.com/measurement-kit/measurement-kit), a C++14 library
implementing many OONI tests.

Note that `export CGO_ENABLED=0` will disable C/C++ extensions and
therefore will prevent Measurement Kit tests from being linked into
the resulting Go binaries. You may want that in some cases, e.g. when
you only want to use OONI tests written in Go.

We plan on gradually rewriting all OONI tests in Go, therefore the
dependency on Measurement Kit will eventually be removed.

Please, read on for platform specific information including how
to install Measurement Kit for each supported platform.

### Android

We don't support Android yet. We'll add support in the future.

### iOS

We don't support iOS yet. We'll add support in the future.

### Linux

We support amd64 only. Create a suitable [Docker])https://www.docker.com/)
container with

```bash
docker build -t gomkbuild .
```

Enter into the development environment with

```bash
docker run -it -v`pwd`:/gomkbuild -w/gomkbuild gomkbuild
```

Then you can use this repository as a normal Go repository.

### macOS

We support amd64 only. Make sure you install our [Homebrew tap](
https://github.com/measurement-kit/homebrew-measurement-kit) and
all the required dependencies with:

```bash
brew tap measurement-kit/measurement-kit
brew install measurement-kit
```

Then you can use this repository as a normal Go repository.

### Windows

We support amd64 only. We cross compile from macOS. Make sure
you install our [Homebrew tap](
https://github.com/measurement-kit/homebrew-measurement-kit) and
all the required dependencies with:

```bash
brew tap measurement-kit/measurement-kit
brew install mingw-w64-measurement-kit
```

Then you can use this repository as a normal Go repository. Make sure you

```bash
export CC=x86_64-w64-mingw32-gcc CXX=x86_64-w64-mingw32-g++ \
  GOOS=windows GOARCH=amd64 CGO_ENABLED=1
```

to produce Windows binaries.

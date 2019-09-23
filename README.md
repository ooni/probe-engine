[![GoDoc](https://godoc.org/github.com/ooni/probe-engine?status.svg)](https://godoc.org/github.com/ooni/probe-engine) [![Build Status](https://travis-ci.org/ooni/probe-engine.svg?branch=master)](https://travis-ci.org/ooni/probe-engine) [![Coverage Status](https://coveralls.io/repos/github/ooni/probe-engine/badge.svg?branch=master)](https://coveralls.io/github/ooni/probe-engine?branch=master) [![Go Report Card](https://goreportcard.com/badge/github.com/ooni/probe-engine)](https://goreportcard.com/report/github.com/ooni/probe-engine)

# OONI probe measurement engine

This repository contains OONI probe's [measurement engine](
https://github.com/ooni/spec/tree/master/probe#engine). That is, the
piece of software that implements OONI nettests.

## API

You can [browse ooni/probe-engine's API](
https://godoc.org/github.com/ooni/probe-engine)
online at godoc.org. We currently don't provide any API
stability guarantees.

This repository also allows to build [miniooni](cmd/miniooni), a
small command line client useful to test the functionality in here
without integrating with OONI probe. You can browse [the manual
of this tool](
https://godoc.org/github.com/ooni/probe-engine/cmd/miniooni)
online at godoc.org. We currently don't promise that the
miniooni CLI will be stable over time.

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

A future version of this document will provide platform specific
instruction for installing Measurement Kit and building.

## Design

See the [design document](DESIGN.md).

## Release procedure

1. make sure that dependencies are up to date

2. make sure that resources are up to date

3. commit, tag, and push

4. create new release on GitHub
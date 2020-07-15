[![GoDoc](https://godoc.org/github.com/ooni/probe-engine?status.svg)](https://godoc.org/github.com/ooni/probe-engine) ![Golang Status](https://github.com/ooni/probe-engine/workflows/golang/badge.svg) ![Android Status](https://github.com/ooni/probe-engine/workflows/android/badge.svg) [![Coverage Status](https://coveralls.io/repos/github/ooni/probe-engine/badge.svg?branch=master)](https://coveralls.io/github/ooni/probe-engine?branch=master) [![Go Report Card](https://goreportcard.com/badge/github.com/ooni/probe-engine)](https://goreportcard.com/report/github.com/ooni/probe-engine)

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
and requires Go v1.14+. You can pull the latest version as a dependency from
your modules aware project by using

```bash
go get -v github.com/ooni/probe-engine
```

## Building miniooni

```bash
go build -v ./cmd/miniooni/
```

## Building Android bindings

```bash
./build-android.bash
```

We automatically build Android bindings whenever commits are pushed to the
`mobile-staging` branch. Such builds could be integrated by using:

```Groovy
maven { url "https://dl.bintray.com/ooni/android/" }
```

and

```Groovy
implementation "org.ooni:oonimkall:VERSION"
```

Where VERSION is like `2020.03.30-231914` corresponding to the moment in
time in which the version has been built.

## Building iOS bindings

```bash
./build-ios.bash
```

We automatically build iOS bindings whenever commits are pushed to the
`mobile-staging` branch. Such builds could be integrated by using:

```ruby
pod 'oonimkall', :podspec => 'https://dl.bintray.com/ooni/ios/oonimkall-VERSION.podspec'
```

Where VERSION is like `2020.03.30-231914` corresponding to the moment in
time in which the version has been built.

## Release procedure

1. make sure that dependencies are up to date

2. make sure that resources are up to date

3. commit, tag, and push

4. create new release on GitHub

## Updating dependencies

1. update direct dependencies using:

```bash
for name in `grep -v indirect go.mod | awk '/^\t/{print $1}'`; do \
  go get -u -v $name;                                             \
done
```

2. pin to a specific psiphon version (we usually track the
`staging-client` branch) using:

```bash
go get -v github.com/Psiphon-Labs/psiphon-tunnel-core@COMMITHASH
```

3. clone `psiphon-tunnel-core`, checkout the tip of the `staging-client`
branch and generate a `go.mod` by running `go mod init && go mod tidy` in
the toplevel dir

4. rewrite `go.mod` such that it contains only your direct dependencies
followed by the exact content of `psiphon-tunnel-core`'s `go.mod`

5. run `go mod tidy`

6. make sure you don't downgrade `bolt` and `goselect` because this
will break downstream builds on MIPS:

```bash
go get -u -v github.com/Psiphon-Labs/bolt github.com/creack/goselect
```

This allows us to pin all psiphon dependencies precisely.

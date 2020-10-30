# OONI probe measurement engine

[![GoDoc](https://godoc.org/github.com/ooni/probe-engine?status.svg)](https://godoc.org/github.com/ooni/probe-engine) [![Short Tests Status](https://github.com/ooni/probe-engine/workflows/shorttests/badge.svg)](https://github.com/ooni/probe-engine/actions?query=workflow%3Ashorttests) [![All Tests Status](https://github.com/ooni/probe-engine/workflows/alltests/badge.svg)](https://github.com/ooni/probe-engine/actions?query=workflow%3Aalltests) [![Coverage Status](https://coveralls.io/repos/github/ooni/probe-engine/badge.svg?branch=master)](https://coveralls.io/github/ooni/probe-engine?branch=master) [![Go Report Card](https://goreportcard.com/badge/github.com/ooni/probe-engine)](https://goreportcard.com/report/github.com/ooni/probe-engine)

This repository contains OONI probe's [measurement engine](
https://github.com/ooni/spec/tree/master/probe#engine). That is, the
piece of software that implements OONI nettests as well as all the
required functionality to run such nettests.

We expect you to use the Go version indicated in [go.mod](go.mod).

## Integrating ooni/probe-engine

We recommend pinning to a specific version of probe-engine:

```bash
go get -v github.com/ooni/probe-engine@VERSION
```

See also the [workflows/using.yml](.github/workflows/using.yml) test
where we check that the latest commit can be imported by a third party.

We do not provide any API stability guarantee.

## Building miniooni

[miniooni](cmd/miniooni) is a small command line client used for
research and quality assurance testing. Build using:

```bash
go build -v ./cmd/miniooni/
```

We don't provide any `miniooni` command line flags stability guarantee.

See

```bash
./miniooni --help
```

for more help.

## Building Android bindings

```bash
./build-android.bash
```

We automatically build Android bindings whenever commits are pushed to the
`mobile-staging` branch. Such builds could be integrated by using:

```Groovy
implementation "org.ooni:oonimkall:VERSION"
```

Where VERSION is like `2020.03.30-231914` corresponding to the
time when the build occurred.

## Building iOS bindings

```bash
./build-ios.bash
```

We automatically build iOS bindings whenever commits are pushed to the
`mobile-staging` branch. Such builds could be integrated by using:

```ruby
pod 'oonimkall', :podspec => 'https://dl.bintray.com/ooni/ios/oonimkall-VERSION.podspec'
```

Where VERSION is like `2020.03.30-231914` corresponding to the
time when the build occurred.

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

The above procedure allows us to pin all psiphon dependencies precisely.

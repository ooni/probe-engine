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
and requires Go v1.13+. We also depend on [Measurement Kit](
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

## Release procedure

1. make sure that dependencies are up to date

2. make sure that resources are up to date

3. commit, tag, and push

4. create new release on GitHub

## Updating dependencies

Updating dependencies is more complex than `go get -u ./...` until our
direct dependency Psiphon migrates to `go.mod`. In turn, they cannot
migrate to `go.mod` until they have support for that in `gomobile`. In
turn `gomobile` with `go.mod` [is in progress](
https://github.com/golang/go/issues/27234). So, we expect to use this
procedure for updating for a few months.

To update direct dependencies use:

```bash
for name in `grep -v indirect go.mod | awk '/^\t/{print $1}'`; do \
  go get -u -v $name;                                             \
done
```

Then update Psiphon. We track the `staging-client` branch. Find the commit
hash of such branch and then run:

```bash
go get -v github.com/Psiphon-Labs/psiphon-tunnel-core@COMMITHASH
```

Then you need to clone `psiphon-tunnel-core` and generate a `go.mod` for
it by running `go mod init && go mod tidy` in its toplevel dir.

Lastly, generate the commands to update using this script:

```Python
import distutils.version
import sys

def slurp(path):
    deps = {}
    with open(path, "r") as filep:
        for line in filep:
            if not line.startswith("\t"):
                continue
            line = line.strip()
            if "// indirect" in line:
                index = line.find("// indirect")
                line = line[:index]
                line = line.strip()
            name, version = line.strip().split()
            deps[name] = version
    return deps

def main():
    if len(sys.argv) != 3:
        sys.exit("usage: %s /our/go.mod /psiphon/go.mod" % sys.argv[0])
    ourdict = slurp(sys.argv[1])
    theirdict = slurp(sys.argv[2])
    for key in theirdict:
        if key not in ourdict:
            continue
        ourver = distutils.version.LooseVersion(ourdict[key])
        theirver = distutils.version.LooseVersion(theirdict[key])
        if theirver <= ourver:
            continue
        print("#", key, theirdict[key], ourdict[key])
        print("go get -v %s@%s" % (key, theirdict[key]))

if __name__ == "__main__":
    main()
```

Run the emitted commands and finally `go mod tidy`.

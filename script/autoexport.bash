#!/bin/bash
set -euxo pipefail

if [[ $# -ne 1 ]]; then
	echo "usage: $0 <tag>" 1>&2
	exit 1
fi
tag=$1
shift

repodir=__repodir__
rm -rf $repodir
git clone -b $tag https://github.com/ooni/probe-cli $repodir

(cd $repodir && git describe --tags > ../UPSTREAM)

rm -rf CODE_OF_CONDUCT.md
cp $repodir/CODE_OF_CONDUCT.md .

rm -rf LICENSE
cp $repodir/LICENSE .

rm -rf go.mod go.sum
cp $repodir/go.mod $repodir/go.sum .

pkgdir=pkg
rm -rf $pkgdir
mv $repodir/internal $pkgdir

GOVERSION=$(cat $repodir/GOVERSION)

rm -rf $repodir

for file in $(find pkg -type f -name \*.go); do
	# See https://stackoverflow.com/a/43190120
	sed $'s|^\t"github.com/ooni/probe-cli/v3/internal|\t"github.com/ooni/probe-engine/pkg|g' $file > $file.new
	mv $file.new $file
done

for file in $(find pkg -type f -name \*.go); do
	sed $'s|^import "github.com/ooni/probe-cli/v3/internal|import "github.com/ooni/probe-engine/pkg|g' $file > $file.new
	mv $file.new $file
done

sed 's|^module github.com/ooni/probe-cli/v3|module github.com/ooni/probe-engine|g' go.mod > go.mod.new
mv go.mod.new go.mod

go mod tidy

echo '# OONI Probe Engine' > README.md
echo '' >> README.md
echo "Automatically exported from github.com/ooni/probe-cli." >> README.md
echo '' >> README.md
echo "Check [UPSTREAM](UPSTREAM) to see the tag/commit from which we exported." >> README.md
echo '' >> README.md
echo 'This is a best effort attempt to export probe-cli internals to community members.' >> README.md
echo '' >> README.md
echo 'We will ignore opened issues and PRs on this repository. You should' >> README.md
echo 'use github.com/ooni/probe and github.com/ooni/probe-cli respectively.' >> README.md

rm -rf .github/workflows
mkdir -p .github/workflows
wf=.github/workflows/build.yml

cat >$wf <<EOF
name: build
on:
  pull_request:
  push:
    branches:
      - "master"
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v3
        with:
          go-version: "$GOVERSION"
      - run: go build ./...
EOF

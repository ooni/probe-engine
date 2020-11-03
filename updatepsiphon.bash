#!/bin/bash
set -ex
basedir=$(cd $(dirname $0) && pwd -P)

oope=github.com/ooni/probe-engine
psirootdir=internal/oopsi
psidir=$psirootdir/github.com/Psiphon-Labs/psiphon-tunnel-core
psitempdir=psi.temp

rm -rf $psirootdir $psitempdir
mkdir -p $psirootdir
git clone -b staging-client https://github.com/Psiphon-Labs/psiphon-tunnel-core $psitempdir

cd $psitempdir
for file in go.mod go.sum; do
  find . -type f -name $file -exec rm -f {} \;
done
for dir in $(find vendor -type d -maxdepth 1 -mindepth 1); do
  mv $dir $basedir/$psirootdir/$(basename $dir)
done
git describe --tags > $basedir/$psirootdir/ooversion.txt
rm -rf vendor .git
cd ..
mv $psitempdir $basedir/$psidir

cd $basedir/$psirootdir
for file in $(find . -type f -name \*.go); do
  cat $file | sed -e "s@github.com/@$oope/$psirootdir/github.com/@g"   \
                  -e "s@go.uber.org/@$oope/$psirootdir/go.uber.org/@g" \
                  -e "s@golang.org/@$oope/$psirootdir/golang.org/@g"   > $file.new
  mv $file.new $file
done

git add .

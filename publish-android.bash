#!/bin/bash
set -e
pkgname=oonimkall
version=$(date -u +%Y%m%dT%H%M%SZ)
baseurl=https://api.bintray.com/content/ooni/android/$pkgname/$version/org/ooni/$pkgname/$version
aarfile=./MOBILE/dist/$pkgname.aar
pomfile=./MOBILE/dist/$pkgname-$version.pom
pomtemplate=./MOBILE/template.pom
user=bassosimone
cat $pomtemplate|sed "s/@VERSION@/$version/g" > $pomfile
if [ -z $BINTRAY_API_KEY ]; then
    echo "FATAL: missing BINTRAY_API_KEY variable" 1>&2
    exit 1
fi
# We currently publish every commit. To cleanup we can fetch all the versions using the
# <curl -s $user:$BINTRAY_API_KEY https://api.bintray.com/packages/ooni/android/oonimkall>
# query, which returns a list of versions. From such list, we can delete the versions we
# don't need using <DELETE /packages/:subject/:repo/:package/versions/:version>.
curl -sT $aarfile -u $user:$BINTRAY_API_KEY $baseurl/$pkgname-$version.aar?publish=1 >/dev/null
curl -sT $pomfile -u $user:$BINTRAY_API_KEY $baseurl/$pkgname-$version.pom?publish=1 >/dev/null
echo "implementation 'org.ooni:oonimkall:$version'"

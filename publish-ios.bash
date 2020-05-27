#!/bin/bash
set -e
pkgname=oonimkall
version=$(date -u +%Y.%m.%d-%H%M%S)
baseurl=https://api.bintray.com/content/ooni/ios/$pkgname/$version/
framework=./MOBILE/dist/$pkgname.framework
frameworkzip=./MOBILE/dist/$pkgname.framework.zip
podspecfile=./MOBILE/dist/$pkgname.podspec
podspectemplate=./MOBILE/template.podspec
user=bassosimone
cat $podspectemplate|sed "s/@VERSION@/$version/g" > $podspecfile
if [ -z $BINTRAY_API_KEY ]; then
    echo "FATAL: missing BINTRAY_API_KEY variable" 1>&2
    exit 1
fi
(cd ./MOBILE/dist && zip $pkgname.framework.zip $pkgname.framework)
# We currently publish every commit. To cleanup we can fetch all the versions using the
# <curl -s $user:$BINTRAY_API_KEY https://api.bintray.com/packages/ooni/android/oonimkall>
# query, which returns a list of versions. From such list, we can delete the versions we
# don't need using <DELETE /packages/:subject/:repo/:package/versions/:version>.
curl -sT $frameworkzip -u $user:$BINTRAY_API_KEY $baseurl/$pkgname.framework.zip?publish=1 >/dev/null
curl -sT $podspecfile -u $user:$BINTRAY_API_KEY $baseurl/$pkgname.podspec?publish=1 >/dev/null
echo "implementation 'org.ooni:oonimkall:$version'"

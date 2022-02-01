#!/usr/bin/env sh

export REPODEST="${GITHUB_WORKSPACE}/packages"
export SRCDEST="${GITHUB_WORKSPACE}/cache/distfiles"
export PACKAGER_PRIVKEY="/root/${INPUT_ABUILD_KEY_NAME}.rsa"
export PACKAGER_PUBKEY="/root/${INPUT_ABUILD_KEY_NAME}.rsa.pub"
printf "${INPUT_ABUILD_KEY}" > "${PACKAGER_PRIVKEY}"
printf "${INPUT_ABUILD_KEY_PUB}" > "${PACKAGER_PUBKEY}"
cp "${PACKAGER_PUBKEY}" /etc/apk/keys/
echo "s/pkgver=.*/pkgver=${INPUT_ABUILD_PKG_VER:1}/"
sed -i "s/pkgver=.*/pkgver=${INPUT_ABUILD_PKG_VER:1}/" APKBUILD
abuild -F checksum
abuild -F -r
find $REPODEST -name "*.apk" | xargs apk verify
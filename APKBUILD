# Contributor: Kalle M. Aagaard <alpine@k-moeller.dk>
# Maintainer: Kalle M. Aagaard <alpine@k-moeller.dk>
pkgname=certmgr
pkgver=0.0.5
pkgrel=1
pkgdesc="A small certapi client"
url="https://github.com/KalleDK/go-certcli/"
arch="all"
license="MIT"
source="${pkgname}-${pkgver}.tar.gz::https://github.com/KalleDK/go-certcli/archive/v${pkgver}.tar.gz"
builddir="${srcdir}/go-${pkgname}-${pkgver}"
makedepends="go"

check() {
    true
}

build() {
    go build -o bin/certcli ./certcli
    mkdir etc
    echo '{"Certs":{}}' > etc/domains.json
}

package() {
    install -Dm755 "bin/certcli" "${pkgdir}/usr/bin/certcli"
    install -Dm750 -d "${pkgdir}/var/lib/certcli"
    install -Dm750 -d "${pkgdir}/var/lib/certcli/certs"
    install -Dm640 "etc/domains.json" "${pkgdir}/var/lib/certcli/domains.json"
}

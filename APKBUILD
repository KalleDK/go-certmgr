# Contributor: Kalle M. Aagaard <alpine@k-moeller.dk>
# Maintainer: Kalle M. Aagaard <alpine@k-moeller.dk>
pkgname=certmgr
pkgver=0.0.6
pkgrel=1
pkgdesc="A small certapi client"
url="https://github.com/KalleDK/go-${pkgname}/"
arch="all"
license="MIT"
source="${pkgname}-${pkgver}.tar.gz::https://github.com/KalleDK/go-${pkgname}/archive/v${pkgver}.tar.gz"
builddir="${srcdir}/go-${pkgname}-${pkgver}"
makedepends="go"

check() {
    true
}

build() {
    go build -o bin/${pkgname} ./${pkgname}

}

package() {
    install -Dm755 "bin/${pkgname}" "${pkgdir}/usr/bin/${pkgname}"
    install -Dm750 -d "${pkgdir}/var/lib/${pkgname}"
    install -Dm750 -d "${pkgdir}/var/lib/${pkgname}/certs"
    install -Dm755 "openrc/${pkgname}" "${pkgdir}/etc/init.d/${pkgname}"
}

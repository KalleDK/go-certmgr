# Contributor: Kalle M. Aagaard <alpine@k-moeller.dk>
# Maintainer: Kalle M. Aagaard <alpine@k-moeller.dk>
pkgname=certmgr
pkgver=0.0.0
pkgrel=1
pkgdesc="A small certmgr service"
url="https://github.com/KalleDK/go-${pkgname}/"
arch="all"
license="MIT"
source="${pkgname}-${pkgver}.tar.gz::https://github.com/KalleDK/go-${pkgname}/archive/v${pkgver}.tar.gz"
builddir="${srcdir}/go-${pkgname}-${pkgver}"
makedepends="go"
subpackages="$pkgname-openrc"

check() {
    true
}

build() {
    go build -o bin/${pkgname} ./cmd/${pkgname}

}

package() {
    install -Dm755 "bin/${pkgname}" "${pkgdir}/usr/bin/${pkgname}"
    install -Dm750 -d "${pkgdir}/var/lib/${pkgname}"
    install -Dm750 -d "${pkgdir}/var/lib/${pkgname}/certs"
    install -Dm755 "openrc/${pkgname}" "${pkgdir}/etc/init.d/${pkgname}"
    install -Dm644 "openrc/conf" "${pkgdir}/etc/conf.d/${pkgname}"
}

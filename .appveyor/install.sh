#!/usr/bin/env bash

set -x
export PATH="/mingw64/bin:${PATH}"
export PKG_CONFIG_PATH="/mingw64/lib/pkgconfig:${PKG_CONFIG_PATH}"
export GOROOT=/mingw64/lib/go
export GOPATH=/go
export CGO_ENABLED=1

curl -O https://repo.msys2.org/msys/x86_64/msys2-keyring-r21.b39fb11-1-any.pkg.tar.xz
curl -O https://repo.msys2.org/msys/x86_64/msys2-keyring-r21.b39fb11-1-any.pkg.tar.xz.sig
pacman-key --verify msys2-keyring-r21.b39fb11-1-any.pkg.tar.xz.sig &&\
pacman --noconfirm -U msys2-keyring-r21.b39fb11-1-any.pkg.tar.xz

pacman -Syy
pacman --noconfirm -S \
    mingw-w64-x86_64-go \
    mingw-w64-x86_64-libusb

find /var/cache/pacman

go version
go get -t github.com/google/gousb/...
go test github.com/google/gousb/...

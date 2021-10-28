#!/usr/bin/env bash

export PATH="/mingw64/bin:${PATH}"
export PKG_CONFIG_PATH="/mingw64/lib/pkgconfig:${PKG_CONFIG_PATH}"
export GOROOT=/mingw64/lib/go
export GOPATH=/go
export CGO_ENABLED=1

set -x

pacman --noconfirm -S \
    mingw-w64-x86_64-go \
    mingw-w64-x86_64-libusb

go version
go get -t github.com/sulfurheron/gousb/...
go test github.com/sulfurheron/gousb/...

#!/usr/bin/env bash

export PATH="/mingw64/bin:${PATH}"
export PKG_CONFIG_PATH="/mingw64/lib/pkgconfig:${PKG_CONFIG_PATH}"
export GOROOT=/mingw64/lib/go
export GOPATH=/go
export CGO_ENABLED=1

pacman --noconfirm -S \
    mingw64/mingw-w64-x86_64-go \
    mingw64/mingw-w64-x86_64-libusb

go version
go get -t github.com/google/gousb/...
go test github.com/google/gousb/...

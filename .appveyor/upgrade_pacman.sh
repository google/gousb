#!/usr/bin/env bash
export PATH="/mingw64/bin:${PATH}"
curl -O https://repo.msys2.org/msys/x86_64/msys2-keyring-r21.b39fb11-1-any.pkg.tar.xz
curl -O https://repo.msys2.org/msys/x86_64/msys2-keyring-r21.b39fb11-1-any.pkg.tar.xz.sig
pacman-key --verify msys2-keyring-r21.b39fb11-1-any.pkg.tar.xz.sig &&\
pacman --noconfirm -U msys2-keyring-r21.b39fb11-1-any.pkg.tar.xz
pacman --noconfirm -Syy pacman

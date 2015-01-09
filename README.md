Introduction
============

The gousb package is an attempt at wrapping the libusb library into a Go-like binding.

Supported platforms include:

- linux
- darwin
- window

Contributing
============
Because I am a Google employee, contributing to this project will require signing the [Google CLA](https://developers.google.com/open-source/cla/individual?csw=1).
This is the same agreement that is required for contributing to Go itself, so if you have
already filled it out for that, you needn't fill it out again.
You will need to send me the email address that you used to sign the agreement
so that I can verify that it is on file before I can accept pull requests.


Installation on Linux
=====================

Dependencies
------------
You must first install [libusb-1.0](http://libusb.org/wiki/libusb-1.0).  This is pretty straightforward on linux and darwin.  The cgo package should be able to find it if you install it in the default manner or use your distribution's package manager.  How to tell cgo how to find one installed in a non-default place is beyond the scope of this README.

*Note*: If you are installing this on darwin, you will probably need to run `fixlibusb_darwin.sh /usr/local/lib/libusb-1.0/libusb.h` because of an LLVM incompatibility.  It shouldn't break C programs, though I haven't tried it in anger.

Example: lsusb
--------------
The gousb project provides a simple but useful example: lsusb.  This binary will list the USB devices connected to your system and various interesting tidbits about them, their configurations, endpoints, etc.  To install it, run the following command:

    go get -v github.com/kylelemons/gousb/lsusb

gousb
-----
If you installed the lsusb example, both libraries below are already installed.

Installing the primary gousb package is really easy:

    go get -v github.com/kylelemons/gousb/usb

There is also a `usbid` package that will not be installed by default by this command, but which provides useful information including the human-readable vendor and product codes for detected hardware.  It's not installed by default and not linked into the `usb` package by default because it adds ~400kb to the resulting binary.  If you want both, they can be installed thus:

    go get -v github.com/kylelemons/gousb/usb{,id}

Installation on Window
=======================
Dependencies
------------
- Gcc (only tested on [Win-Builds](http://win-builds.org/), but any MSYS, CYGWIN, MINGW should be worked
- [libusb-1.0](http://sourceforge.net/projects/libusb/files/libusb-1.0/)

Build
-----

- After downloaded, extract them to some directory; such as D:\lib\libusb-1.0.xx\
- Remember two path which "libusb.h" file and "libusb-1.0.a" inside

*Note* For MinGW32, use **MinGW32/static/libusb-1.0.a** while MinGW64 use **MinGW64/static/libusb-1.0.a** for linker

- Open `$(GOPATH)/src/github.com/.../gousb/usb/usb.go`. Then inject:

        // #cgo CFLAGS: -ID:/lib/libusbx-1.0.xx/include
        // #cgo LDFLAGS: D:/lib/libusbx-1.0.xx/MinGW64/static/libusb-1.0.a


before the line:

    // #include <libusb-1.0/libusb.h>

This flag will tell the linker the exact path of static library

- Go to `$(GOPATH)/src/github.com/.../gousb/usb/`. Run:

        $ go install

    
- Go to `$(GOPATH)/src/github.com/.../gousb/usbid/`. Run:

        $ go install

Example: lsusb
--------------
- Go to `$(GOPATH)/src/github.com/.../gousb/lsusb/`. Run:

        $ go run main.go

Documentation
=============
The documentation can be viewed via local godoc or via Gary Burd's excellent [godoc.org](http://godoc.org/):

- [usb](http://godoc.org/github.com/kylelemons/gousb/usb)
- [usbid](http://godoc.org/pkg/github.com/kylelemons/gousb/usbid)

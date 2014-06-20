Introduction
============

The gousb package is an attempt at wrapping the libusb library into a Go-like binding.

Supported platforms include:

- linux
- darwin

Windows support is unconfirmed, but should work via cgo and [libusb-win32](http://sourceforge.net/apps/trac/libusb-win32/wiki).

Contributing
============
Because I am a Google employee, contributing to this project will require signing the [Google CLA](https://developers.google.com/open-source/cla/individual?csw=1).
This is the same agreement that is required for contributing to Go itself, so if you have
already filled it out for that, you needn't fill it out again.
You will need to send me the email address that you used to sign the agreement
so that I can verify that it is on file before I can accept pull requests.


Installation
============

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

Documentation
=============
The documentation can be viewed via local godoc or via Gary Burd's excellent [godoc.org](http://godoc.org/):

- [usb](http://godoc.org/github.com/kylelemons/gousb/usb)
- [usbid](http://godoc.org/pkg/github.com/kylelemons/gousb/usbid)

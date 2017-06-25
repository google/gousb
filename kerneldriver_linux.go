// +build linux

// Licensed under the Apache License, Version 2.0 (the "License");
// File originally made by jpoirier (https://github.com/jpoirier)
// Addapted to the google/gousb library by nicovell3

package gousb

import (
	"fmt"
)

/*
#cgo pkg-config: libusb-1.0
#include <libusb.h>
*/
import "C"

// detachKernelDriver detaches any active kernel drivers, if supported by the platform.
// If there are any errors, like Context.ListDevices, only the final one will be returned.
func (d *Device) detachKernelDriver() (err error) {
	for _, cfg := range d.Desc.Configs {
		for _, iface := range cfg.Interfaces {
			switch activeErr := C.libusb_kernel_driver_active((*C.libusb_device_handle)(d.handle), C.int(iface.Number)); activeErr {
			case C.LIBUSB_ERROR_NOT_SUPPORTED:
				// no need to do any futher checking, no platform support
				return
			case 0:
				continue
			case 1:
				switch detachErr := C.libusb_detach_kernel_driver((*C.libusb_device_handle)(d.handle), C.int(iface.Number)); detachErr {
				case C.LIBUSB_ERROR_NOT_SUPPORTED:
					// shouldn't ever get here, should be caught by the outer switch
					return
				case 0:
					d.detached[iface.Number]=true
				case C.LIBUSB_ERROR_NOT_FOUND:
					// this status is returned if libusb's driver is already attached to the device
					d.detached[iface.Number]=true
				default:
					return fmt.Errorf("Can't detach kernel driver: %v", detachErr)
				}
			default:
				return fmt.Errorf("Active kernel driver check: %v", activeErr)
			}
		}
	}

	return
}

// attachKernelDriver re-attaches kernel drivers to any previously detached interfaces, if supported by the platform.
// If there are any errors, like Context.ListDevices, only the final one will be returned.
func (d *Device) attachKernelDriver() (err error) {
	for iface := range d.detached {
		switch attachErr := C.libusb_attach_kernel_driver((*C.libusb_device_handle)(d.handle), C.int(iface)); attachErr {
		case C.LIBUSB_ERROR_NOT_SUPPORTED:
			// no need to do any futher checking, no platform support
			return
		case 0:
			continue
		default:
			return fmt.Errorf("Can't reattach kernel driver: %v", attachErr)
		}
	}

	return
} 

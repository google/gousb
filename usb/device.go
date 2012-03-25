package usb

// #cgo LDFLAGS: -lusb-1.0
// #include <libusb-1.0/libusb.h>
import "C"

import (
	"log"
	"runtime"
)

type DeviceInfo struct {
	dev *C.libusb_device
}

func newDeviceInfo(dev *C.libusb_device) *DeviceInfo {
	d := &DeviceInfo{
		dev: dev,
	}

	// Reference this device
	C.libusb_ref_device(dev)

	// I still can't get this to be called
	runtime.SetFinalizer(d, (*DeviceInfo).Close)

	log.Printf("deviceInfo %p initialized", d.dev)
	return d
}

// Get the assigned Bus Number for this USB device.
func (d *DeviceInfo) BusNumber() byte {
	return byte(C.libusb_get_bus_number(d.dev))
}

// Get the assigned Bus Address for this USB device.
func (d *DeviceInfo) Address() byte {
	return byte(C.libusb_get_device_address(d.dev))
}

// Get a descriptor for the device.
func (d *DeviceInfo) Descriptor() (*Descriptor, error) {
	return newDescriptor(d.dev)
}

// Open the given device for I/O.
func (d *DeviceInfo) Open() (*Device, error) {
	var handle *C.libusb_device_handle
	if errno := C.libusb_open(d.dev, &handle); errno != 0 {
		return nil, usbError(errno)
	}

	return newDevice(handle), nil
}

// Close decrements the reference count for the device in the libusb driver
// code.  It should be called exactly once!
func (d *DeviceInfo) Close() error {
	if d.dev != nil {
		log.Printf("deviceInfo %p closed", d.dev)
		C.libusb_unref_device(d.dev)
	}
	d.dev = nil
	return nil
}

type Device struct {
	handle *C.libusb_device_handle
}

func newDevice(handle *C.libusb_device_handle) *Device {
	d := &Device{
		handle: handle,
	}

	// :(
	runtime.SetFinalizer(d, (*Device).Close)

	log.Printf("device %p initialized", d.handle)
	return d
}

func (d *Device) Close() error {
	if d.handle != nil {
		log.Printf("device %p closed", d.handle)
		C.libusb_unref_device(d.handle)
	}
	d.handle = nil
	return nil
}

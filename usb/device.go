package usb

// #cgo LDFLAGS: -lusb-1.0
// #include <libusb-1.0/libusb.h>
import "C"

import (
	"fmt"
	"reflect"
	"runtime"
	"time"
	"unsafe"
)

var DefaultControlTimeout = 5 * time.Second

type Device struct {
	handle *C.libusb_device_handle

	// Embed the device information for easy access
	*Descriptor

	// Timeouts
	ControlTimeout time.Duration
}

func newDevice(handle *C.libusb_device_handle, desc *Descriptor) *Device {
	d := &Device{
		handle:         handle,
		Descriptor:     desc,
		ControlTimeout: DefaultControlTimeout,
	}

	// This doesn't seem to actually get called
	runtime.SetFinalizer(d, (*Device).Close)

	return d
}

func (d *Device) Reset() error {
	if errno := C.libusb_reset_device(d.handle); errno != 0 {
		return usbError(errno)
	}
	return nil
}

func (d *Device) Control(rType, request uint8, val, idx uint16, data []byte) (int, error) {
	dataSlice := (*reflect.SliceHeader)(unsafe.Pointer(&data))
	n := C.libusb_control_transfer(
		d.handle,
		C.uint8_t(rType),
		C.uint8_t(request),
		C.uint16_t(val),
		C.uint16_t(idx),
		(*C.uchar)(unsafe.Pointer(dataSlice.Data)),
		C.uint16_t(len(data)),
		C.uint(d.ControlTimeout/time.Millisecond))
	if n < 0 {
		return int(n), usbError(n)
	}
	return int(n), nil
}

// ActiveConfig returns the config id (not the index) of the active configuration.
// This corresponds to the ConfigInfo.Config field.
func (d *Device) ActiveConfig() (uint8, error) {
	var cfg C.int
	if errno := C.libusb_get_configuration(d.handle, &cfg); errno < 0 {
		return 0, usbError(errno)
	}
	return uint8(cfg), nil
}

// SetConfig attempts to change the active configuration.
// The cfg provided is the config id (not the index) of the configuration to set,
// which corresponds to the ConfigInfo.Config field.
func (d *Device) SetConfig(cfg uint8) error {
	if errno := C.libusb_set_configuration(d.handle, C.int(cfg)); errno < 0 {
		return usbError(errno)
	}
	return nil
}

// Close the device.
func (d *Device) Close() error {
	if d.handle == nil {
		return fmt.Errorf("usb: double close on device")
	}
	C.libusb_close(d.handle)
	d.handle = nil
	return nil
}

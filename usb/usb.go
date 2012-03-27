package usb

// #cgo LDFLAGS: -lusb-1.0
// #include <libusb-1.0/libusb.h>
import "C"

import (
	"reflect"
	"runtime"
	"unsafe"
)

type Context struct {
	ctx *C.libusb_context
}

func (c *Context) Debug(level int) {
	C.libusb_set_debug(c.ctx, C.int(level))
}

func NewContext() *Context {
	c := new(Context)

	//log.Printf("gousb initialized")
	if errno := C.libusb_init(&c.ctx); errno != 0 {
		panic(usbError(errno))
	}

	// This doesn't seem to actually get called.  Sigh.
	runtime.SetFinalizer(c, (*Context).Close)

	return c
}

// ListDevices calls each with each enumerated device.  If the function returns
// true, the device is added to the list of *DeviceInfo to return.  All of
// these must be Closed, even if an error is also returned.
func (c *Context) ListDevices(each func(bus, addr int, desc *Descriptor) bool) ([]*DeviceInfo, error) {
	var list **C.libusb_device
	cnt := C.libusb_get_device_list(c.ctx, &list)
	if cnt < 0 {
		return nil, usbError(cnt)
	}
	defer C.libusb_free_device_list(list, 1)

	var slice []*C.libusb_device
	*(*reflect.SliceHeader)(unsafe.Pointer(&slice)) = reflect.SliceHeader{
		Data: uintptr(unsafe.Pointer(list)),
		Len:  int(cnt),
		Cap:  int(cnt),
	}

	ret := []*DeviceInfo{}
	for _, dev := range slice {
		bus := int(C.libusb_get_bus_number(dev))
		addr := int(C.libusb_get_device_address(dev))
		desc, err := newDescriptor(dev)
		if err != nil {
			return ret, err
		}
		if each(bus, addr, desc) {
			ret = append(ret, newDeviceInfo(dev))
		}
	}
	return ret, nil
}

func (c *Context) Close() error {
	if c.ctx != nil {
		C.libusb_exit(c.ctx)
		//log.Printf("gousb finished")
	}
	c.ctx = nil
	return nil
}

package usb

// #cgo LDFLAGS: -lusb-1.0
// #include <libusb-1.0/libusb.h>
import "C"

import (
	"log"
	"reflect"
	"runtime"
	"unsafe"
)

type Endpoint struct {
	Type          DescriptorType
	Address       ID
	MaxPacketSize uint16
	PollInterval  uint8
	RefreshRate   uint8
	SynchAddress  ID
}

type Interface struct {
	Type       DescriptorType
	Number     ID
	Alternate  ID
	IfClass    Class
	IfSubClass uint8
	IfProtocol uint8
	Endpoints  []*Endpoint
}

type Config struct {
	cfg *C.struct_libusb_config_descriptor

	Type       DescriptorType
	Config     ID
	Attributes uint8
	MaxPower   uint8
	Interfaces [][]*Interface
}

func newConfig(cfg *C.struct_libusb_config_descriptor) *Config {
	c := &Config{
		cfg:        cfg,
		Type:       DescriptorType(cfg.bDescriptorType),
		Config:     ID(cfg.bConfigurationValue),
		Attributes: uint8(cfg.bmAttributes),
		MaxPower:   uint8(cfg.MaxPower),
	}

	var ifaces []C.struct_libusb_interface
	*(*reflect.SliceHeader)(unsafe.Pointer(&ifaces)) = reflect.SliceHeader{
		Data: uintptr(unsafe.Pointer(cfg._interface)),
		Len:  int(cfg.bNumInterfaces),
		Cap:  int(cfg.bNumInterfaces),
	}
	c.Interfaces = make([][]*Interface, 0, len(ifaces))
	for _, iface := range ifaces {
		var alts []C.struct_libusb_interface_descriptor
		*(*reflect.SliceHeader)(unsafe.Pointer(&alts)) = reflect.SliceHeader{
			Data: uintptr(unsafe.Pointer(iface.altsetting)),
			Len:  int(iface.num_altsetting),
			Cap:  int(iface.num_altsetting),
		}
		descs := make([]*Interface, 0, len(alts))
		for _, alt := range alts {
			i := &Interface{
				Type:       DescriptorType(alt.bDescriptorType),
				Number:     ID(alt.bInterfaceNumber),
				Alternate:  ID(alt.bAlternateSetting),
				IfClass:    Class(alt.bInterfaceClass),
				IfSubClass: uint8(alt.bInterfaceSubClass),
				IfProtocol: uint8(alt.bInterfaceProtocol),
			}
			var ends []C.struct_libusb_endpoint_descriptor
			*(*reflect.SliceHeader)(unsafe.Pointer(&ends)) = reflect.SliceHeader{
				Data: uintptr(unsafe.Pointer(alt.endpoint)),
				Len:  int(alt.bNumEndpoints),
				Cap:  int(alt.bNumEndpoints),
			}
			i.Endpoints = make([]*Endpoint, 0, len(ends))
			for _, end := range ends {
				i.Endpoints = append(i.Endpoints, &Endpoint{
					Type:          DescriptorType(end.bDescriptorType),
					Address:       ID(end.bEndpointAddress),
					MaxPacketSize: uint16(end.wMaxPacketSize),
					PollInterval:  uint8(end.bInterval),
					RefreshRate:   uint8(end.bRefresh),
					SynchAddress:  ID(end.bSynchAddress),
				})
			}
			descs = append(descs, i)
		}
		c.Interfaces = append(c.Interfaces, descs)
	}

	// *sigh*
	runtime.SetFinalizer(c, (*Config).Close)

	log.Printf("config %p initialized", c.cfg)
	return c
}

// Close decrements the reference count for the device in the libusb driver
// code.  It should be called exactly once!
func (c *Config) Close() error {
	if c.cfg != nil {
		log.Printf("config %p closed", c.cfg)
		C.libusb_free_config_descriptor(c.cfg)
	}
	c.cfg = nil
	return nil
}

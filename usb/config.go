package usb

// #cgo LDFLAGS: -lusb-1.0
// #include <libusb-1.0/libusb.h>
import "C"

import (
	"fmt"
	"reflect"
	"unsafe"
)

type EndpointInfo struct {
	Type          DescriptorType
	Address       uint8
	Attributes    uint8
	MaxPacketSize uint16
	PollInterval  uint8
	RefreshRate   uint8
	SynchAddress  uint8
}

func (e EndpointInfo) Number() int {
	return int(e.Address) & ENDPOINT_NUM_MASK
}

func (e EndpointInfo) Direction() EndpointDirection {
	return EndpointDirection(e.Address) & ENDPOINT_DIR_MASK
}

func (e EndpointInfo) String() string {
	return fmt.Sprintf("Endpoint %d %-3s %s - %s %s",
		e.Number(), e.Direction(),
		TransferType(e.Attributes)&TRANSFER_TYPE_MASK,
		IsoSyncType(e.Attributes)&ISO_SYNC_TYPE_MASK,
		IsoUsageType(e.Attributes)&ISO_USAGE_TYPE_MASK,
	)
}

type InterfaceInfo struct {
	Type       DescriptorType
	Number     uint8
	Alternate  uint8
	IfClass    uint8
	IfSubClass uint8
	IfProtocol uint8
	Endpoints  []EndpointInfo
}

func (i InterfaceInfo) String() string {
	return fmt.Sprintf("Interface %02x (config %02x)", i.Number, i.Alternate)
}

type ConfigInfo struct {
	Type       DescriptorType
	Config     uint8
	Attributes uint8
	MaxPower   uint8
	Interfaces [][]InterfaceInfo
}

func (c ConfigInfo) String() string {
	return fmt.Sprintf("Config %02x", c.Config)
}

func newConfig(cfg *C.struct_libusb_config_descriptor) ConfigInfo {
	c := ConfigInfo{
		Type:       DescriptorType(cfg.bDescriptorType),
		Config:     uint8(cfg.bConfigurationValue),
		Attributes: uint8(cfg.bmAttributes),
		MaxPower:   uint8(cfg.MaxPower),
	}

	var ifaces []C.struct_libusb_interface
	*(*reflect.SliceHeader)(unsafe.Pointer(&ifaces)) = reflect.SliceHeader{
		Data: uintptr(unsafe.Pointer(cfg._interface)),
		Len:  int(cfg.bNumInterfaces),
		Cap:  int(cfg.bNumInterfaces),
	}
	c.Interfaces = make([][]InterfaceInfo, 0, len(ifaces))
	for _, iface := range ifaces {
		var alts []C.struct_libusb_interface_descriptor
		*(*reflect.SliceHeader)(unsafe.Pointer(&alts)) = reflect.SliceHeader{
			Data: uintptr(unsafe.Pointer(iface.altsetting)),
			Len:  int(iface.num_altsetting),
			Cap:  int(iface.num_altsetting),
		}
		descs := make([]InterfaceInfo, 0, len(alts))
		for _, alt := range alts {
			i := InterfaceInfo{
				Type:       DescriptorType(alt.bDescriptorType),
				Number:     uint8(alt.bInterfaceNumber),
				Alternate:  uint8(alt.bAlternateSetting),
				IfClass:    uint8(alt.bInterfaceClass),
				IfSubClass: uint8(alt.bInterfaceSubClass),
				IfProtocol: uint8(alt.bInterfaceProtocol),
			}
			var ends []C.struct_libusb_endpoint_descriptor
			*(*reflect.SliceHeader)(unsafe.Pointer(&ends)) = reflect.SliceHeader{
				Data: uintptr(unsafe.Pointer(alt.endpoint)),
				Len:  int(alt.bNumEndpoints),
				Cap:  int(alt.bNumEndpoints),
			}
			i.Endpoints = make([]EndpointInfo, 0, len(ends))
			for _, end := range ends {
				i.Endpoints = append(i.Endpoints, EndpointInfo{
					Type:          DescriptorType(end.bDescriptorType),
					Address:       uint8(end.bEndpointAddress),
					Attributes:    uint8(end.bmAttributes),
					MaxPacketSize: uint16(end.wMaxPacketSize),
					PollInterval:  uint8(end.bInterval),
					RefreshRate:   uint8(end.bRefresh),
					SynchAddress:  uint8(end.bSynchAddress),
				})
			}
			descs = append(descs, i)
		}
		c.Interfaces = append(c.Interfaces, descs)
	}
	return c
}

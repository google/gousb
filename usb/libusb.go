// Copyright 2017 the gousb Authors.  All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package usb

import (
	"fmt"
	"log"
	"reflect"
	"time"
	"unsafe"
)

/*
#include <libusb.h>

int compact_iso_data(struct libusb_transfer *xfer, unsigned char *status);
int submit(struct libusb_transfer *xfer);
*/
import "C"

type libusbContext C.libusb_context
type libusbDevice C.libusb_device
type libusbDevHandle C.libusb_device_handle
type libusbTransfer C.struct_libusb_transfer
type libusbIso C.struct_libusb_iso_packet_descriptor

// libusbIntf is a set of trivial idiomatic Go wrappers around libusb C functions.
// The underlying code is generally not testable or difficult to test,
// since libusb interacts directly with the host USB stack.
//
// All functions here should operate on types defined on C.libusb* data types,
// and occasionally on convenience data types (like TransferType or Descriptor).
type libusbIntf interface {
	// context
	init() (*libusbContext, error)
	handleEvents(*libusbContext, <-chan struct{})
	getDevices(*libusbContext) ([]*libusbDevice, error)
	exit(*libusbContext)
	setDebug(*libusbContext, int)
	openVIDPID(*libusbContext, int, int) (*libusbDevice, *libusbDevHandle, error)

	// device
	dereference(*libusbDevice)
	getDeviceDesc(*libusbDevice) (*Descriptor, error)
	open(*libusbDevice) (*libusbDevHandle, error)

	close(*libusbDevHandle)
	reset(*libusbDevHandle) error
	control(*libusbDevHandle, time.Duration, uint8, uint8, uint16, uint16, []byte) (int, error)
	getConfig(*libusbDevHandle) (uint8, error)
	setConfig(*libusbDevHandle, uint8) error
	getStringDesc(*libusbDevHandle, int) (string, error)
	setAutoDetach(*libusbDevHandle, int) error

	// interface
	claim(*libusbDevHandle, uint8) error
	release(*libusbDevHandle, uint8)
	setAlt(*libusbDevHandle, uint8, uint8) error

	// endpoint
	alloc(*libusbDevHandle, int) (*libusbTransfer, error)
	cancel(*libusbTransfer) error
	submit(*libusbTransfer) error
	free(*libusbTransfer)
	setIsoPacketLengths(*libusbTransfer, uint32)
	compactIsoData(*libusbTransfer) (int, TransferStatus)
}

// libusbImpl is an implementation of libusbIntf using real CGo-wrapped libusb.
type libusbImpl struct{}

func (libusbImpl) init() (*libusbContext, error) {
	var ctx *C.libusb_context
	if err := fromUSBError(C.libusb_init(&ctx)); err != nil {
		return nil, err
	}
	return (*libusbContext)(ctx), nil
}

func (libusbImpl) handleEvents(c *libusbContext, done <-chan struct{}) {
	tv := C.struct_timeval{tv_sec: 10}
	for {
		select {
		case <-done:
			return
		default:
		}
		if errno := C.libusb_handle_events_timeout_completed((*C.libusb_context)(c), &tv, nil); errno < 0 {
			log.Printf("handle_events: error: %s", usbError(errno))
		}
	}
}

func (libusbImpl) getDevices(ctx *libusbContext) ([]*libusbDevice, error) {
	var list **C.libusb_device
	cnt := C.libusb_get_device_list((*C.libusb_context)(ctx), &list)
	if cnt < 0 {
		return nil, fromUSBError(C.int(cnt))
	}
	var devs []*C.libusb_device
	*(*reflect.SliceHeader)(unsafe.Pointer(&devs)) = reflect.SliceHeader{
		Data: uintptr(unsafe.Pointer(list)),
		Len:  int(cnt),
		Cap:  int(cnt),
	}
	var ret []*libusbDevice
	for _, d := range devs {
		ret = append(ret, (*libusbDevice)(d))
	}
	// devices will be dereferenced later, during close.
	C.libusb_free_device_list(list, 0)
	return ret, nil
}

func (libusbImpl) exit(c *libusbContext) {
	C.libusb_exit((*C.libusb_context)(c))
}

func (libusbImpl) setDebug(c *libusbContext, lvl int) {
	C.libusb_set_debug((*C.libusb_context)(c), C.int(lvl))
}

func (libusbImpl) openVIDPID(ctx *libusbContext, vid, pid int) (*libusbDevice, *libusbDevHandle, error) {
	h := C.libusb_open_device_with_vid_pid((*C.libusb_context)(ctx), (C.uint16_t)(vid), (C.uint16_t)(pid))
	if h == nil {
		return nil, nil, ERROR_NOT_FOUND
	}
	dev := C.libusb_get_device(h)
	if dev == nil {
		return nil, nil, ERROR_NO_DEVICE
	}
	C.libusb_ref_device(dev)
	return (*libusbDevice)(dev), (*libusbDevHandle)(h), nil
}

func (libusbImpl) getDeviceDesc(d *libusbDevice) (*Descriptor, error) {
	var desc C.struct_libusb_device_descriptor
	if err := fromUSBError(C.libusb_get_device_descriptor((*C.libusb_device)(d), &desc)); err != nil {
		return nil, err
	}
	// Enumerate configurations
	var cfgs []ConfigInfo
	for i := 0; i < int(desc.bNumConfigurations); i++ {
		var cfg *C.struct_libusb_config_descriptor
		if err := fromUSBError(C.libusb_get_config_descriptor((*C.libusb_device)(d), C.uint8_t(i), &cfg)); err != nil {
			return nil, err
		}
		c := ConfigInfo{
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
		c.Interfaces = make([]InterfaceInfo, 0, len(ifaces))
		for _, iface := range ifaces {
			if iface.num_altsetting == 0 {
				continue
			}

			var alts []C.struct_libusb_interface_descriptor
			*(*reflect.SliceHeader)(unsafe.Pointer(&alts)) = reflect.SliceHeader{
				Data: uintptr(unsafe.Pointer(iface.altsetting)),
				Len:  int(iface.num_altsetting),
				Cap:  int(iface.num_altsetting),
			}
			descs := make([]InterfaceSetup, 0, len(alts))
			for _, alt := range alts {
				i := InterfaceSetup{
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
					ei := EndpointInfo{
						Address:       uint8(end.bEndpointAddress),
						Attributes:    uint8(end.bmAttributes),
						MaxPacketSize: uint16(end.wMaxPacketSize),
						PollInterval:  uint8(end.bInterval),
						RefreshRate:   uint8(end.bRefresh),
						SynchAddress:  uint8(end.bSynchAddress),
					}
					if ei.TransferType() == TRANSFER_TYPE_ISOCHRONOUS {
						// bits 0-10 identify the packet size, bits 11-12 are the number of additional transactions per microframe.
						// Don't use libusb_get_max_iso_packet_size, as it has a bug where it returns the same value
						// regardless of alternative setting used, where different alternative settings might define different
						// max packet sizes.
						// See http://libusb.org/ticket/77 for more background.
						ei.MaxIsoPacket = uint32(end.wMaxPacketSize) & 0x07ff * (uint32(end.wMaxPacketSize)>>11&3 + 1)
					}
					i.Endpoints = append(i.Endpoints, ei)
				}
				descs = append(descs, i)
			}
			c.Interfaces = append(c.Interfaces, InterfaceInfo{
				Number: descs[0].Number,
				Setups: descs,
			})
		}
		C.libusb_free_config_descriptor(cfg)
		cfgs = append(cfgs, c)
	}

	return &Descriptor{
		Bus:      uint8(C.libusb_get_bus_number((*C.libusb_device)(d))),
		Address:  uint8(C.libusb_get_device_address((*C.libusb_device)(d))),
		Spec:     BCD(desc.bcdUSB),
		Device:   BCD(desc.bcdDevice),
		Vendor:   ID(desc.idVendor),
		Product:  ID(desc.idProduct),
		Class:    uint8(desc.bDeviceClass),
		SubClass: uint8(desc.bDeviceSubClass),
		Protocol: uint8(desc.bDeviceProtocol),
		Configs:  cfgs,
	}, nil
}

func (libusbImpl) dereference(d *libusbDevice) {
	C.libusb_unref_device((*C.libusb_device)(d))
}

func (libusbImpl) open(d *libusbDevice) (*libusbDevHandle, error) {
	var handle *C.libusb_device_handle
	if err := fromUSBError(C.libusb_open((*C.libusb_device)(d), &handle)); err != nil {
		return nil, err
	}
	return (*libusbDevHandle)(handle), nil
}

func (libusbImpl) close(d *libusbDevHandle) {
	C.libusb_close((*C.libusb_device_handle)(d))
}

func (libusbImpl) reset(d *libusbDevHandle) error {
	return fromUSBError(C.libusb_reset_device((*C.libusb_device_handle)(d)))
}

func (libusbImpl) control(d *libusbDevHandle, timeout time.Duration, rType, request uint8, val, idx uint16, data []byte) (int, error) {
	dataSlice := (*reflect.SliceHeader)(unsafe.Pointer(&data))
	n := C.libusb_control_transfer(
		(*C.libusb_device_handle)(d),
		C.uint8_t(rType),
		C.uint8_t(request),
		C.uint16_t(val),
		C.uint16_t(idx),
		(*C.uchar)(unsafe.Pointer(dataSlice.Data)),
		C.uint16_t(len(data)),
		C.uint(timeout/time.Millisecond))
	if n < 0 {
		return int(n), fromUSBError(n)
	}
	return int(n), nil
}

func (libusbImpl) getConfig(d *libusbDevHandle) (uint8, error) {
	var cfg C.int
	if errno := C.libusb_get_configuration((*C.libusb_device_handle)(d), &cfg); errno < 0 {
		return 0, fromUSBError(errno)
	}
	return uint8(cfg), nil
}

func (libusbImpl) setConfig(d *libusbDevHandle, cfg uint8) error {
	return fromUSBError(C.libusb_set_configuration((*C.libusb_device_handle)(d), C.int(cfg)))
}

func (libusbImpl) getStringDesc(d *libusbDevHandle, index int) (string, error) {
	// allocate 200-byte array limited the length of string descriptor
	buf := make([]byte, 200)
	// get string descriptor from libusb. if errno < 0 then there are any errors.
	// if errno >= 0; it is a length of result string descriptor
	errno := C.libusb_get_string_descriptor_ascii(
		(*C.libusb_device_handle)(d),
		C.uint8_t(index),
		(*C.uchar)(unsafe.Pointer(&buf[0])),
		200)
	if errno < 0 {
		return "", fmt.Errorf("usb: getstr: %s", fromUSBError(errno))
	}
	return string(buf[:errno]), nil
}

func (libusbImpl) setAutoDetach(d *libusbDevHandle, val int) error {
	err := fromUSBError(C.libusb_set_auto_detach_kernel_driver((*C.libusb_device_handle)(d), C.int(val)))
	if err != nil && err != ERROR_NOT_SUPPORTED {
		return err
	}
	return nil
}

func (libusbImpl) claim(d *libusbDevHandle, iface uint8) error {
	return fromUSBError(C.libusb_claim_interface((*C.libusb_device_handle)(d), C.int(iface)))
}

func (libusbImpl) release(d *libusbDevHandle, iface uint8) {
	C.libusb_release_interface((*C.libusb_device_handle)(d), C.int(iface))
}

func (libusbImpl) setAlt(d *libusbDevHandle, iface, setup uint8) error {
	return fromUSBError(C.libusb_set_interface_alt_setting((*C.libusb_device_handle)(d), C.int(iface), C.int(setup)))
}

func (libusbImpl) alloc(d *libusbDevHandle, isoPackets int) (*libusbTransfer, error) {
	xfer := C.libusb_alloc_transfer(C.int(isoPackets))
	if xfer == nil {
		return nil, fmt.Errorf("libusb_alloc_transfer(%d) failed", isoPackets)
	}
	xfer.dev_handle = (*C.libusb_device_handle)(d)
	xfer.num_iso_packets = C.int(isoPackets)
	return (*libusbTransfer)(xfer), nil
}

func (libusbImpl) cancel(t *libusbTransfer) error {
	return fromUSBError(C.libusb_cancel_transfer((*C.struct_libusb_transfer)(t)))
}

func (libusbImpl) submit(t *libusbTransfer) error {
	return fromUSBError(C.submit((*C.struct_libusb_transfer)(t)))
}

func (libusbImpl) free(t *libusbTransfer) {
	C.libusb_free_transfer((*C.struct_libusb_transfer)(t))
}

func (libusbImpl) setIsoPacketLengths(t *libusbTransfer, length uint32) {
	C.libusb_set_iso_packet_lengths((*C.struct_libusb_transfer)(t), C.uint(length))
}

func (libusbImpl) compactIsoData(t *libusbTransfer) (int, TransferStatus) {
	var status TransferStatus
	n := int(C.compact_iso_data((*C.struct_libusb_transfer)(t), (*C.uchar)(unsafe.Pointer(&status))))
	return n, status
}

// libusb is an injection point for tests
var libusb libusbIntf = libusbImpl{}

var (
	libusbIsoSize   = C.sizeof_struct_libusb_iso_packet_descriptor
	libusbIsoOffset = unsafe.Offsetof(C.struct_libusb_transfer{}.iso_packet_desc)
)

//export xfer_callback
func xfer_callback(cptr unsafe.Pointer) {
	ch := *(*chan struct{})(cptr)
	close(ch)
}

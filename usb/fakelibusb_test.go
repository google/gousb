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
	"sync"
	"time"
	"unsafe"
)

var (
	fakeDevices = map[*libusbDevice]*Descriptor{
		// Bus 001 Device 001: ID 9999:0001
		// One config, one interface, one setup,
		// two endpoints: 0x01 OUT, 0x82 IN.
		// TODO(https://github.com/golang/go/issues/19487): replace with
		// new(libusbDevice)
		(*libusbDevice)(unsafe.Pointer(uintptr(1))): &Descriptor{
			Bus:      1,
			Address:  1,
			Spec:     USB_2_0,
			Device:   BCD(0x0100), // 1.00
			Vendor:   ID(0x9999),
			Product:  ID(0x0001),
			Protocol: 255,
			Configs: []ConfigInfo{{
				Config:   1,
				MaxPower: 50, // * 2mA
				Interfaces: []InterfaceInfo{{
					Number: 0,
					Setups: []InterfaceSetup{{
						Number:    0,
						Alternate: 0,
						IfClass:   uint8(CLASS_VENDOR_SPEC),
						Endpoints: []EndpointInfo{{
							Address:       uint8(0x01 | ENDPOINT_DIR_OUT),
							Attributes:    uint8(TRANSFER_TYPE_BULK),
							MaxPacketSize: 512,
						}, {
							Address:       uint8(0x02 | ENDPOINT_DIR_IN),
							Attributes:    uint8(TRANSFER_TYPE_BULK),
							MaxPacketSize: 512,
						}},
					}},
				}},
			}},
		},
		// Bus 001 Device 002: ID 8888:0002
		// One config, two interfaces. interface #0 with no endpoints,
		// interface #1 with two alt setups with different packet sizes for
		// endpoints. Two isochronous endpoints, 0x05 OUT and 0x86 OUT.
		// TODO(https://github.com/golang/go/issues/19487): replace with
		// new(libusbDevice)
		(*libusbDevice)(unsafe.Pointer(uintptr(2))): &Descriptor{
			Bus:      1,
			Address:  2,
			Spec:     USB_2_0,
			Device:   BCD(0x0103), // 1.03
			Vendor:   ID(0x8888),
			Product:  ID(0x0002),
			Protocol: 255,
			Configs: []ConfigInfo{{
				Config:   1,
				MaxPower: 50, // * 2mA
				Interfaces: []InterfaceInfo{{
					Number: 0,
					Setups: []InterfaceSetup{{
						Number:    0,
						Alternate: 0,
						IfClass:   uint8(CLASS_VENDOR_SPEC),
					}},
				}, {
					Number: 1,
					Setups: []InterfaceSetup{{
						Number:    1,
						Alternate: 0,
						IfClass:   uint8(CLASS_VENDOR_SPEC),
						Endpoints: []EndpointInfo{{
							Address:       uint8(0x05 | ENDPOINT_DIR_OUT),
							Attributes:    uint8(TRANSFER_TYPE_ISOCHRONOUS),
							MaxPacketSize: 2<<11 | 1024,
							MaxIsoPacket:  3 * 1024,
						}, {
							Address:       uint8(0x06 | ENDPOINT_DIR_IN),
							Attributes:    uint8(TRANSFER_TYPE_ISOCHRONOUS),
							MaxPacketSize: 2<<11 | 1024,
							MaxIsoPacket:  3 * 1024,
						}},
					}, {
						Number:    1,
						Alternate: 1,
						IfClass:   uint8(CLASS_VENDOR_SPEC),
						Endpoints: []EndpointInfo{{
							Address:       uint8(0x05 | ENDPOINT_DIR_OUT),
							Attributes:    uint8(TRANSFER_TYPE_ISOCHRONOUS),
							MaxPacketSize: 1<<11 | 1024,
							MaxIsoPacket:  2 * 1024,
						}, {
							Address:       uint8(0x06 | ENDPOINT_DIR_IN),
							Attributes:    uint8(TRANSFER_TYPE_ISOCHRONOUS),
							MaxPacketSize: 1<<11 | 1024,
							MaxIsoPacket:  2 * 1024,
						}},
					}, {
						Number:    1,
						Alternate: 2,
						IfClass:   uint8(CLASS_VENDOR_SPEC),
						Endpoints: []EndpointInfo{{
							Address:       uint8(0x05 | ENDPOINT_DIR_OUT),
							Attributes:    uint8(TRANSFER_TYPE_ISOCHRONOUS),
							MaxPacketSize: 1024,
							MaxIsoPacket:  1024,
						}, {
							Address:       uint8(0x06 | ENDPOINT_DIR_IN),
							Attributes:    uint8(TRANSFER_TYPE_ISOCHRONOUS),
							MaxPacketSize: 1024,
							MaxIsoPacket:  1024,
						}},
					}},
				}},
			}},
		},
	}
)

type fakeTransfer struct {
	// done is the channel that needs to be closed when the transfer has finished.
	done chan struct{}
	// buf is the slice for reading/writing data between the submit() and wait() returning.
	buf []byte
	// status will be returned by wait() on this transfer
	status TransferStatus
	// length is the number of bytes used from the buffer (write) or available
	// in the buffer (read).
	length int
}

type fakeLibusb struct {
	libusbIntf

	mu sync.Mutex
	// ts has a map of all allocated transfers, indexed by the pointer of
	// underlying libusbTransfer.
	ts map[*libusbTransfer]*fakeTransfer
	// submitted receives a fakeTransfers when submit() is called.
	submitted chan *fakeTransfer
	// handlers is a map of device handles pointing at opened devices.
	handlers map[*libusbDevHandle]*libusbDevice
}

func (f *fakeLibusb) alloc(_ *libusbDevHandle, _ uint8, _ TransferType, _ time.Duration, _ int, buf []byte) (*libusbTransfer, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	t := new(libusbTransfer)
	f.ts[t] = &fakeTransfer{buf: buf}
	return t, nil
}

func (f *fakeLibusb) submit(t *libusbTransfer, done chan struct{}) error {
	f.mu.Lock()
	ft := f.ts[t]
	f.mu.Unlock()
	ft.done = done
	f.submitted <- ft
	return nil
}

func (f *fakeLibusb) cancel(t *libusbTransfer) error { return nil }
func (f *fakeLibusb) free(t *libusbTransfer) {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.ts, t)
}

func (f *fakeLibusb) data(t *libusbTransfer) (int, TransferStatus) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.ts[t].length, f.ts[t].status
}

func (f *fakeLibusb) waitForSubmitted() *fakeTransfer {
	return <-f.submitted
}

func (f *fakeLibusb) getDevices(*libusbContext) ([]*libusbDevice, error) {
	ret := make([]*libusbDevice, 0, len(fakeDevices))
	for d := range fakeDevices {
		ret = append(ret, d)
	}
	return ret, nil
}

func (f *fakeLibusb) getDeviceDesc(d *libusbDevice) (*Descriptor, error) {
	if desc, ok := fakeDevices[d]; ok {
		return desc, nil
	}
	return nil, fmt.Errorf("invalid USB device %p", d)
}

func (f *fakeLibusb) dereference(d *libusbDevice) {}

func (f *fakeLibusb) open(d *libusbDevice) (*libusbDevHandle, error) {
	h := new(libusbDevHandle)
	f.mu.Lock()
	defer f.mu.Unlock()
	f.handlers[h] = d
	return h, nil
}

func (f *fakeLibusb) close(h *libusbDevHandle) {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.handlers, h)
}

func newFakeLibusb() *fakeLibusb {
	return &fakeLibusb{
		ts:         make(map[*libusbTransfer]*fakeTransfer),
		submitted:  make(chan *fakeTransfer, 10),
		handlers:   make(map[*libusbDevHandle]*libusbDevice),
		libusbIntf: libusbImpl{},
	}
}

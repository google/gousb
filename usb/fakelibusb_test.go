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
	"errors"
	"fmt"
	"sync"
	"time"
	"unsafe"
)

var (
	// fake devices connected through the fakeLibusb stack.
	fakeDevices = []*Descriptor{
		// Bus 001 Device 001: ID 9999:0001
		// One config, one interface, one setup,
		// two endpoints: 0x01 OUT, 0x82 IN.
		&Descriptor{
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
		&Descriptor{
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

type fakeDevice struct {
	desc *Descriptor
	alt  uint8
}

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

// fakeLibusb implements a fake libusb stack that pretends to have a number of
// devices connected to it (see fakeDevices variable for a list of devices).
// fakeLibusb is expected to implement all the functions related to device
// enumeration, configuration etc. according to fakeDevices descriptors.
// The fake devices endpoints don't have any particular behavior implemented,
// instead fakeLibusb provides additional functions, like waitForSubmitted,
// that allows the test to explicitly control individual transfer behavior.
type fakeLibusb struct {
	mu sync.Mutex
	// fakeDevices has a map of devices and their descriptors.
	fakeDevices map[*libusbDevice]*fakeDevice
	// ts has a map of all allocated transfers, indexed by the pointer of
	// underlying libusbTransfer.
	ts map[*libusbTransfer]*fakeTransfer
	// submitted receives a fakeTransfers when submit() is called.
	submitted chan *fakeTransfer
	// handles is a map of device handles pointing at opened devices.
	handles map[*libusbDevHandle]*libusbDevice
	// claims is a map of devices to a set of claimed interfaces
	claims map[*libusbDevice]map[uint8]bool
}

func (f *fakeLibusb) init() (*libusbContext, error)                       { return new(libusbContext), nil }
func (f *fakeLibusb) handleEvents(c *libusbContext, done <-chan struct{}) { <-done }
func (f *fakeLibusb) getDevices(*libusbContext) ([]*libusbDevice, error) {
	ret := make([]*libusbDevice, 0, len(fakeDevices))
	for d := range f.fakeDevices {
		ret = append(ret, d)
	}
	return ret, nil
}
func (f *fakeLibusb) exit(*libusbContext)          {}
func (f *fakeLibusb) setDebug(*libusbContext, int) {}

func (f *fakeLibusb) dereference(d *libusbDevice) {}
func (f *fakeLibusb) getDeviceDesc(d *libusbDevice) (*Descriptor, error) {
	if dev, ok := f.fakeDevices[d]; ok {
		return dev.desc, nil
	}
	return nil, fmt.Errorf("invalid USB device %p", d)
}
func (f *fakeLibusb) open(d *libusbDevice) (*libusbDevHandle, error) {
	h := new(libusbDevHandle)
	f.mu.Lock()
	defer f.mu.Unlock()
	f.handles[h] = d
	return h, nil
}

func (f *fakeLibusb) close(h *libusbDevHandle) {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.handles, h)
}
func (f *fakeLibusb) reset(*libusbDevHandle) error { return nil }
func (f *fakeLibusb) control(*libusbDevHandle, time.Duration, uint8, uint8, uint16, uint16, []byte) (int, error) {
	return 0, errors.New("not implemented")
}
func (f *fakeLibusb) getConfig(*libusbDevHandle) (uint8, error) { return 1, nil }
func (f *fakeLibusb) setConfig(d *libusbDevHandle, cfg uint8) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if len(f.claims[f.handles[d]]) != 0 {
		return fmt.Errorf("can't set device config while interfaces are claimed: %v", f.claims[f.handles[d]])
	}
	if cfg != 1 {
		return fmt.Errorf("device doesn't have config number %d", cfg)
	}
	return nil
}
func (f *fakeLibusb) getStringDesc(*libusbDevHandle, int) (string, error) {
	return "", errors.New("not implemented")
}
func (f *fakeLibusb) setAutoDetach(*libusbDevHandle, int) error { return nil }

func (f *fakeLibusb) claim(d *libusbDevHandle, intf uint8) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	c := f.claims[f.handles[d]]
	if c == nil {
		c = make(map[uint8]bool)
		f.claims[f.handles[d]] = c
	}
	c[intf] = true
	return nil
}
func (f *fakeLibusb) release(d *libusbDevHandle, intf uint8) {
	f.mu.Lock()
	defer f.mu.Unlock()
	c := f.claims[f.handles[d]]
	if c == nil {
		return
	}
	c[intf] = false
}
func (f *fakeLibusb) setAlt(d *libusbDevHandle, intf, alt uint8) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if !f.claims[f.handles[d]][intf] {
		return fmt.Errorf("interface %d must be claimed before alt setup can be set", intf)
	}
	f.fakeDevices[f.handles[d]].alt = alt
	return nil
}

func (f *fakeLibusb) alloc(_ *libusbDevHandle, _ uint8, _ TransferType, _ time.Duration, _ int, buf []byte) (*libusbTransfer, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	t := new(libusbTransfer)
	f.ts[t] = &fakeTransfer{buf: buf}
	return t, nil
}
func (f *fakeLibusb) cancel(t *libusbTransfer) error { return errors.New("not implemented") }
func (f *fakeLibusb) submit(t *libusbTransfer, done chan struct{}) error {
	f.mu.Lock()
	ft := f.ts[t]
	f.mu.Unlock()
	ft.done = done
	f.submitted <- ft
	return nil
}
func (f *fakeLibusb) data(t *libusbTransfer) (int, TransferStatus) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.ts[t].length, f.ts[t].status
}
func (f *fakeLibusb) free(t *libusbTransfer) {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.ts, t)
}
func (f *fakeLibusb) setIsoPacketLengths(*libusbTransfer, uint32) {}

// waitForSubmitted can be used by tests to define custom behavior of the transfers submitted on the USB bus.
// TODO(sebek): add fields in fakeTransfer to differentiate between different devices/endpoints used concurrently.
func (f *fakeLibusb) waitForSubmitted() *fakeTransfer {
	return <-f.submitted
}

func newFakeLibusb() *fakeLibusb {
	fl := &fakeLibusb{
		fakeDevices: make(map[*libusbDevice]*fakeDevice),
		ts:          make(map[*libusbTransfer]*fakeTransfer),
		submitted:   make(chan *fakeTransfer, 10),
		handles:     make(map[*libusbDevHandle]*libusbDevice),
		claims:      make(map[*libusbDevice]map[uint8]bool),
	}
	for i, d := range fakeDevices {
		// libusb does not export a way to allocate a new libusb_device struct
		// without using the full USB stack. Since the fake library uses the
		// libusbDevice only as an identifier, use arbitrary numbers pretending
		// to be pointers. The contents of these pointers is never accessed.
		fl.fakeDevices[(*libusbDevice)(unsafe.Pointer(uintptr(i)))] = &fakeDevice{
			desc: d,
			alt:  0,
		}
	}
	return fl
}

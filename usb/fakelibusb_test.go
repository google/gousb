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
	"sync"
	"time"
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

func newFakeLibusb() *fakeLibusb {
	return &fakeLibusb{
		ts:         make(map[*libusbTransfer]*fakeTransfer),
		submitted:  make(chan *fakeTransfer, 10),
		handlers:   make(map[*libusbDevHandle]*libusbDevice),
		libusbIntf: libusbImpl{},
	}
}

var (
	fakeDev1 = new(libusbDevice)
	fakeDev2 = new(libusbDevice)
)

func (f *fakeLibusb) getDevices(*libusbContext) ([]*libusbDevice, error) {
	return []*libusbDevice{fakeDev1, fakeDev2}, nil
}

func (f *fakeLibusb) getDeviceDesc(d *libusbDevice) (*Descriptor, error) {
	switch d {
	case fakeDev1:
		// Bus 001 Device 001: ID 9999:0001
		// One config, one interface, one setup,
		// two endpoints: 0x01 write, 0x82 read.
		return &Descriptor{
			Bus:      1,
			Address:  1,
			Spec:     USB_2_0,
			Device:   BCD(0x0100), // 1.00
			Vendor:   ID(0x9999),
			Product:  ID(0x0001),
			Protocol: 255,
			Configs: []ConfigInfo{{
				Config:     1,
				Attributes: uint8(TRANSFER_TYPE_BULK),
				MaxPower:   50, // * 2mA
				Interfaces: []InterfaceInfo{{
					Number: 0,
					Setups: []InterfaceSetup{{
						Number:    0,
						Alternate: 0,
						IfClass:   uint8(CLASS_VENDOR_SPEC),
						Endpoints: []EndpointInfo{},
					}},
				}},
			}},
		}, nil
	case fakeDev2:
		return &Descriptor{}, nil
	}
	return nil, errors.New("invalid USB device")
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

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

package gousb

import (
	"errors"
	"fmt"
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
func (f *fakeLibusb) getDeviceDesc(d *libusbDevice) (*DeviceDesc, error) {
	if dev, ok := f.fakeDevices[d]; ok {
		return dev.devDesc, nil
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
	debug.Printf("setConfig(%p, %d)\n", d, cfg)
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
func (f *fakeLibusb) getStringDesc(d *libusbDevHandle, index int) (string, error) {
	dev, ok := f.fakeDevices[f.handles[d]]
	if !ok {
		return "", fmt.Errorf("invalid USB device %p", d)
	}
	str, ok := dev.strDesc[index]
	if !ok {
		return "", fmt.Errorf("invalid string descriptor index %d", index)
	}
	return str, nil
}
func (f *fakeLibusb) setAutoDetach(*libusbDevHandle, int) error { return nil }

func (f *fakeLibusb) detachKernelDriver(*libusbDevHandle, uint8) error { return nil }

func (f *fakeLibusb) claim(d *libusbDevHandle, intf uint8) error {
	debug.Printf("claim(%p, %d)\n", d, intf)
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
	debug.Printf("release(%p, %d)\n", d, intf)
	f.mu.Lock()
	defer f.mu.Unlock()
	c := f.claims[f.handles[d]]
	if c == nil {
		return
	}
	c[intf] = false
}
func (f *fakeLibusb) setAlt(d *libusbDevHandle, intf, alt uint8) error {
	debug.Printf("setAlt(%p, %d, %d)\n", d, intf, alt)
	f.mu.Lock()
	defer f.mu.Unlock()
	if !f.claims[f.handles[d]][intf] {
		return fmt.Errorf("interface %d must be claimed before alt setup can be set", intf)
	}
	f.fakeDevices[f.handles[d]].alt = alt
	return nil
}

func (f *fakeLibusb) alloc(_ *libusbDevHandle, _ *EndpointDesc, _ time.Duration, _ int, bufLen int, done chan struct{}) (*libusbTransfer, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	t := newFakeTransferPointer()
	f.ts[t] = &fakeTransfer{buf: make([]byte, bufLen), done: done}
	return t, nil
}
func (f *fakeLibusb) cancel(t *libusbTransfer) error { return errors.New("not implemented") }
func (f *fakeLibusb) submit(t *libusbTransfer) error {
	f.mu.Lock()
	ft := f.ts[t]
	f.mu.Unlock()
	f.submitted <- ft
	return nil
}
func (f *fakeLibusb) buffer(t *libusbTransfer) []byte { return f.ts[t].buf }
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

// empty can be used to confirm that all transfers were cleaned up.
func (f *fakeLibusb) empty() bool {
	return len(f.submitted) == 0
}

func newFakeLibusb() (*fakeLibusb, func() error) {
	origLibusb := libusb
	fl := &fakeLibusb{
		fakeDevices: make(map[*libusbDevice]*fakeDevice),
		ts:          make(map[*libusbTransfer]*fakeTransfer),
		submitted:   make(chan *fakeTransfer, 10),
		handles:     make(map[*libusbDevHandle]*libusbDevice),
		claims:      make(map[*libusbDevice]map[uint8]bool),
	}
	for _, d := range fakeDevices {
		// libusb does not export a way to allocate a new libusb_device struct
		// without using the full USB stack. Since the fake library uses the
		// libusbDevice only as an identifier, use an arbitrary unique pointer.
		// The contents of these pointers is never accessed.
		fd := new(fakeDevice)
		*fd = d
		fl.fakeDevices[newDevicePointer()] = fd
	}
	libusb = fl
	return fl, func() error {
		defer func() { libusb = origLibusb }()
		close(fl.submitted)
		if got := len(fl.ts); got > 0 {
			for t := range fl.ts {
				fl.free(t)
			}
			return fmt.Errorf("fakeLibusb has %d remaining transfers that should have been freed", got)
		}
		return nil
	}
}

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
		libusbIntf: libusbImpl{},
	}
}

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

import "sync"

type fakeLibusb struct {
	sync.Mutex
	ts map[*libusbTransfer]chan struct{}
	libusbIntf
}

func (f *fakeLibusb) submit(t *libusbTransfer) error {
	f.Lock()
	defer f.Unlock()
	if f.ts[t] == nil {
		f.ts[t] = make(chan struct{})
	}
	close(f.ts[t])
	return nil
}

func (f *fakeLibusb) cancel(t *libusbTransfer) error { return nil }

func (f *fakeLibusb) waitForSubmit(t *usbTransfer) {
	f.Lock()
	if f.ts[t.xfer] == nil {
		f.ts[t.xfer] = make(chan struct{})
	}
	ch := f.ts[t.xfer]
	f.Unlock()
	<-ch
}

func (f *fakeLibusb) runCallback(t *usbTransfer, cb func(*usbTransfer)) {
	f.Lock()
	defer f.Unlock()
	delete(f.ts, t.xfer)
	cb(t)
	close(t.done)
}

func newFakeLibusb() *fakeLibusb {
	return &fakeLibusb{
		ts:         make(map[*libusbTransfer]chan struct{}),
		libusbIntf: libusbImpl{},
	}
}

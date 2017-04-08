// Copyright 2013 Google Inc.  All rights reserved.
// Copyright 2016 the gousb Authors.  All rights reserved.
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
	"time"
)

type Endpoint struct {
	h *libusbDevHandle

	InterfaceSetup
	Info EndpointInfo

	readTimeout  time.Duration
	writeTimeout time.Duration
}

func (e *Endpoint) String() string {
	return e.Info.String()
}

func (e *Endpoint) Read(buf []byte) (int, error) {
	if e.Info.Direction != EndpointDirectionIn {
		return 0, fmt.Errorf("usb: read: not an IN endpoint")
	}

	return e.transfer(buf, e.readTimeout)
}

func (e *Endpoint) Write(buf []byte) (int, error) {
	if e.Info.Direction != EndpointDirectionOut {
		return 0, fmt.Errorf("usb: write: not an OUT endpoint")
	}

	return e.transfer(buf, e.writeTimeout)
}

func (e *Endpoint) transfer(buf []byte, timeout time.Duration) (int, error) {
	if len(buf) == 0 {
		return 0, nil
	}

	t, err := newUSBTransfer(e.h, &e.Info, buf, timeout)
	if err != nil {
		return 0, err
	}
	defer t.free()

	if err := t.submit(); err != nil {
		return 0, err
	}

	n, err := t.wait()
	if err != nil {
		return n, err
	}
	return n, nil
}

func newEndpoint(h *libusbDevHandle, s InterfaceSetup, e EndpointInfo, rt, wt time.Duration) *Endpoint {
	return &Endpoint{
		InterfaceSetup: s,
		Info:           e,
		h:              h,
		readTimeout:    rt,
		writeTimeout:   wt,
	}
}

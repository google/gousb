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

type Endpoint interface {
	Read(b []byte) (int, error)
	Write(b []byte) (int, error)
	Interface() InterfaceSetup
	Info() EndpointInfo
}

type endpoint struct {
	h *libusbDevHandle

	InterfaceSetup
	EndpointInfo

	readTimeout  time.Duration
	writeTimeout time.Duration
}

func (e *endpoint) Read(buf []byte) (int, error) {
	if EndpointDirection(e.Address)&ENDPOINT_DIR_MASK != ENDPOINT_DIR_IN {
		return 0, fmt.Errorf("usb: read: not an IN endpoint")
	}

	return e.transfer(buf, e.readTimeout)
}

func (e *endpoint) Write(buf []byte) (int, error) {
	if EndpointDirection(e.Address)&ENDPOINT_DIR_MASK != ENDPOINT_DIR_OUT {
		return 0, fmt.Errorf("usb: write: not an OUT endpoint")
	}

	return e.transfer(buf, e.writeTimeout)
}

func (e *endpoint) Interface() InterfaceSetup { return e.InterfaceSetup }
func (e *endpoint) Info() EndpointInfo        { return e.EndpointInfo }

func (e *endpoint) transfer(buf []byte, timeout time.Duration) (int, error) {
	if len(buf) == 0 {
		return 0, nil
	}

	t, err := newUSBTransfer(e.h, e.EndpointInfo, buf, timeout)
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

func newEndpoint(h *libusbDevHandle, s InterfaceSetup, e EndpointInfo, rt, wt time.Duration) *endpoint {
	return &endpoint{
		InterfaceSetup: s,
		EndpointInfo:   e,
		h:              h,
		readTimeout:    rt,
		writeTimeout:   wt,
	}
}

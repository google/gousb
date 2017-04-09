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
	"strings"
	"time"
)

// EndpointInfo contains the information about an interface endpoint, extracted
// from the descriptor.
type EndpointInfo struct {
	// Number represents the endpoint number. Note that the endpoint number is different from the
	// address field in the descriptor - address 0x82 means endpoint number 2,
	// with endpoint direction IN.
	// The device can have up to two endpoints with the same number but with
	// different directions.
	Number uint8
	// Direction defines whether the data is flowing IN or OUT from the host perspective.
	Direction EndpointDirection
	// MaxPacketSize is the maximum USB packet size for a single frame/microframe.
	MaxPacketSize uint32
	// TransferType defines the endpoint type - bulk, interrupt, isochronous.
	TransferType TransferType
	// PollInterval is the maximum time between transfers for interrupt and isochronous transfer,
	// or the NAK interval for a control transfer. See endpoint descriptor bInterval documentation
	// in the USB spec for details.
	PollInterval time.Duration
	// IsoSyncType is the isochronous endpoint synchronization type, as defined by USB spec.
	IsoSyncType IsoSyncType
	// UsageType is the isochronous or interrupt endpoint usage type, as defined by USB spec.
	UsageType UsageType
}

func endpointAddr(n uint8, d EndpointDirection) uint8 {
	addr := n
	if d == EndpointDirectionIn {
		addr |= 0x80
	}
	return addr
}

// String returns the human-readable description of the endpoint.
func (e EndpointInfo) String() string {
	ret := make([]string, 0, 3)
	ret = append(ret, fmt.Sprintf("Endpoint #%d %s (address 0x%02x) %s", e.Number, e.Direction, endpointAddr(e.Number, e.Direction), e.TransferType))
	switch e.TransferType {
	case TransferTypeIsochronous:
		ret = append(ret, fmt.Sprintf("- %s %s", e.IsoSyncType, e.UsageType))
	case TransferTypeInterrupt:
		ret = append(ret, fmt.Sprintf("- %s", e.UsageType))
	}
	ret = append(ret, fmt.Sprintf("[%d bytes]", e.MaxPacketSize))
	return strings.Join(ret, " ")
}

// Endpoint identifies a USB endpoint opened for transfer.
type Endpoint struct {
	h *libusbDevHandle

	InterfaceSetting
	Info EndpointInfo

	readTimeout  time.Duration
	writeTimeout time.Duration
}

// String returns a human-readable description of the endpoint.
func (e *Endpoint) String() string {
	return e.Info.String()
}

// Read reads data from an IN endpoint.
func (e *Endpoint) Read(buf []byte) (int, error) {
	if e.Info.Direction != EndpointDirectionIn {
		return 0, fmt.Errorf("usb: read: not an IN endpoint")
	}

	return e.transfer(buf, e.readTimeout)
}

// Write writes data to an OUT endpoint.
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

func newEndpoint(h *libusbDevHandle, s InterfaceSetting, e EndpointInfo, rt, wt time.Duration) *Endpoint {
	return &Endpoint{
		InterfaceSetting: s,
		Info:             e,
		h:                h,
		readTimeout:      rt,
		writeTimeout:     wt,
	}
}

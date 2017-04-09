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

// InterfaceInfo contains information about a USB interface, extracted from
// the descriptor.
type InterfaceInfo struct {
	// Number is the number of this interface, a zero-based index in the array
	// of interfaces supported by the device configuration.
	Number uint8
	// AltSettings is a list of alternate settings supported by the interface.
	AltSettings []InterfaceSetting
}

// String returns a human-readable descripton of the interface and it's
// alternate settings.
func (i InterfaceInfo) String() string {
	return fmt.Sprintf("Interface %d (%d alternate settings)", i.Number, len(i.AltSettings))
}

// InterfaceSetting contains information about a USB interface with a particular
// alternate setting, extracted from the descriptor.
type InterfaceSetting struct {
	// Number is the number of this interface, the same as in InterfaceInfo.
	Number uint8
	// Alternate is the number of this alternate setting.
	Alternate uint8
	// Class is the USB-IF class code, as defined by the USB spec.
	Class Class
	// SubClass is the USB-IF subclass code, as defined by the USB spec.
	SubClass Class
	// Protocol is USB protocol code, as defined by the USB spe.c
	Protocol Protocol
	// Endpoints has the list of endpoints available on this interface with
	// this alternate setting.
	Endpoints []EndpointInfo
}

// String returns a human-readable descripton of the particular
// alternate setting of an interface.
func (a InterfaceSetting) String() string {
	return fmt.Sprintf("Interface %d alternate setting %d", a.Number, a.Alternate)
}

// ConfigInfo contains the information about a USB device configuration.
type ConfigInfo struct {
	// Config is the configuration number.
	Config uint8
	// SelfPowered is true if the device is powered externally, i.e. not
	// drawing power from the USB bus.
	SelfPowered bool
	// RemoteWakeup is true if the device supports remote wakeup.
	RemoteWakeup bool
	// MaxPower is the maximum current the device draws from the USB bus
	// in this configuration.
	MaxPower Milliamperes
	// Interfaces has a list of USB interfaces available in this configuration.
	Interfaces []InterfaceInfo
}

// String returns the human-readable description of the configuration.
func (c ConfigInfo) String() string {
	return fmt.Sprintf("Config %d", c.Config)
}

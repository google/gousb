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

import "fmt"

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

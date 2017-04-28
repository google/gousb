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
)

// InterfaceInfo contains information about a USB interface, extracted from
// the descriptor.
type InterfaceInfo struct {
	// Number is the number of this interface, a zero-based index in the array
	// of interfaces supported by the device configuration.
	Number int
	// AltSettings is a list of alternate settings supported by the interface.
	AltSettings []InterfaceSetting
}

// String returns a human-readable descripton of the interface descriptor and
// it's alternate settings.
func (i InterfaceInfo) String() string {
	return fmt.Sprintf("Interface %d (%d alternate settings)", i.Number, len(i.AltSettings))
}

// InterfaceSetting contains information about a USB interface with a particular
// alternate setting, extracted from the descriptor.
type InterfaceSetting struct {
	// Number is the number of this interface, the same as in InterfaceInfo.
	Number int
	// Alternate is the number of this alternate setting.
	Alternate int
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

type Interface struct {
	Setting InterfaceSetting

	config *Config
}

func (i *Interface) String() string {
	return fmt.Sprintf("%s,if=%d,alt=%d", i.config, i.Setting.Number, i.Setting.Alternate)
}

// Close releases the interface.
func (i *Interface) Close() {
	if i.config == nil {
		return
	}
	libusb.release(i.config.dev.handle, uint8(i.Setting.Number))
	i.config.mu.Lock()
	defer i.config.mu.Unlock()
	delete(i.config.claimed, i.Setting.Number)
	i.config = nil
}

func (i *Interface) openEndpoint(epNum int) (*endpoint, error) {
	var ep EndpointInfo
	var found bool
	for _, e := range i.Setting.Endpoints {
		if e.Number == epNum {
			debug.Printf("found ep %s in %s\n", e, i)
			ep = e
			found = true
			break
		}
	}
	if !found {
		return nil, fmt.Errorf("%s does not have endpoint number %d. Available endpoints: %v", i, epNum, i.Setting.Endpoints)
	}
	return &endpoint{
		InterfaceSetting: i.Setting,
		Info:             ep,
		h:                i.config.dev.handle,
	}, nil
}

// InEndpoint prepares an IN endpoint for transfer.
func (i *Interface) InEndpoint(epNum int) (*InEndpoint, error) {
	if i.config == nil {
		return nil, fmt.Errorf("InEndpoint(%d) called on %s after Close", epNum, i)
	}
	ep, err := i.openEndpoint(epNum)
	if err != nil {
		return nil, err
	}
	if ep.Info.Direction != EndpointDirectionIn {
		return nil, fmt.Errorf("%s is not an IN endpoint", ep)
	}
	return &InEndpoint{
		endpoint: ep,
	}, nil
}

// OutEndpoint prepares an OUT endpoint for transfer.
func (i *Interface) OutEndpoint(epNum int) (*OutEndpoint, error) {
	if i.config == nil {
		return nil, fmt.Errorf("OutEndpoint(%d) called on %s after Close", epNum, i)
	}
	ep, err := i.openEndpoint(epNum)
	if err != nil {
		return nil, err
	}
	if ep.Info.Direction != EndpointDirectionOut {
		return nil, fmt.Errorf("%s is not an OUT endpoint", ep)
	}
	return &OutEndpoint{
		endpoint: ep,
	}, nil
}

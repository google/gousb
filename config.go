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

package gousb

import (
	"fmt"
	"sync"
)

// ConfigInfo contains the information about a USB device configuration,
// extracted from the device descriptor.
type ConfigDesc struct {
	// Config is the configuration number.
	Config int
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

// String returns the human-readable description of the configuration descriptor.
func (c ConfigDesc) String() string {
	return fmt.Sprintf("Configuration %d", c.Config)
}

// Config represents a USB device set to use a particular configuration.
// Only one Config of a particular device can be used at any one time.
// To access device endpoints, claim an interface and it's alternate
// setting number through a call to Interface().
type Config struct {
	Info ConfigDesc

	dev *Device

	// Claimed interfaces
	mu      sync.Mutex
	claimed map[int]bool
}

// Close releases the underlying device, allowing the caller to switch the device to a different configuration.
func (c *Config) Close() error {
	if c.dev == nil {
		return nil
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.claimed) > 0 {
		var ifs []int
		for k := range c.claimed {
			ifs = append(ifs, k)
		}
		return fmt.Errorf("failed to release %s, interfaces %v are still open", c, ifs)
	}
	c.dev.mu.Lock()
	defer c.dev.mu.Unlock()
	c.dev.claimed = nil
	c.dev = nil
	return nil
}

// String returns the human-readable description of the configuration.
func (c *Config) String() string {
	return fmt.Sprintf("%s,config=%d", c.dev.String(), c.Info.Config)
}

// Interface claims and returns an interface on a USB device.
// inft specifies the number of an interface to claim, and alt specifies the
// alternate setting number for that interface.
func (c *Config) Interface(intf, alt int) (*Interface, error) {
	if c.dev == nil {
		return nil, fmt.Errorf("Interface(%d, %d) called on %s after Close", intf, alt, c)
	}
	if intf < 0 || intf >= len(c.Info.Interfaces) {
		return nil, fmt.Errorf("interface %d not found in %s, available interfaces 0..%d", intf, c, len(c.Info.Interfaces)-1)
	}
	ifInfo := c.Info.Interfaces[intf]
	if alt < 0 || alt >= len(ifInfo.AltSettings) {
		return nil, fmt.Errorf("alternate setting %d not found for %s in %s, available alt settings 0..%d", alt, ifInfo, c, len(ifInfo.AltSettings)-1)
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	if c.claimed[intf] {
		return nil, fmt.Errorf("interface %d on %s is already claimed", intf, c)
	}

	// Claim the interface
	if err := libusb.claim(c.dev.handle, uint8(intf)); err != nil {
		return nil, fmt.Errorf("failed to claim interface %d on %s: %v", intf, c, err)
	}

	if err := libusb.setAlt(c.dev.handle, uint8(intf), uint8(alt)); err != nil {
		libusb.release(c.dev.handle, uint8(intf))
		return nil, fmt.Errorf("failed to set alternate config %d on interface %d of %s: %v", alt, intf, c, err)
	}

	c.claimed[intf] = true
	return &Interface{
		Setting: ifInfo.AltSettings[alt],
		config:  c,
	}, nil
}
